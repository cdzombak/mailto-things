package main

import (
	"context"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"log"
	"mime"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"google.golang.org/api/gmail/v1"
)

const (
	envVarConfigDir            = "MAILTO_THINGS_CONFIG_DIR"
	envVarIncomingEmail        = "MAILTO_THINGS_INCOMING_EMAIL"
	envVarOutgoingEmail        = "MAILTO_THINGS_OUTGOING_EMAIL"
	envVarAttachmentsDir       = "MAILTO_THINGS_ATTACHMENTS_DIR"
	envVarAttachmentsDirURL    = "MAILTO_THINGS_ATTACHMENTS_DIR_URL"
	envVarDontTouchOrigMessage = "MAILTO_THINGS_DONT_TOUCH_ORIG_MESSAGE"
)

var (
	configDirFlag         = flag.String("configDir", "", "Path to the directory where Gmail app credentials & user tokens are stored. Overrides environment variable MAILTO_THINGS_CONFIG_DIR.")
	attachmentsDirFlag    = flag.String("attachmentsDir", "", "Path to the directory where attachments are stored. Overrides environment variable MAILTO_THINGS_ATTACHMENTS_DIR.")
	attachmentsDirURLFlag = flag.String("attachmentsDirURL", "", "URL to the directory where attachments are stored. Should not end with a slash. Overrides environment variable MAILTO_THINGS_ATTACHMENTS_DIR_URL.")
	incomingEmailFlag     = flag.String("incomingEmail", "", "Email address which receives tasks with attachments. Overrides environment variable MAILTO_THINGS_INCOMING_EMAIL.")
	outgoingEmailFlag     = flag.String("outgoingEmail", "", "Things email address to send task emails to. Overrides environment variable MAILTO_THINGS_OUTGOING_EMAIL.")
	fileCreateModeFlag    = flag.String("fileCreateMode", "0600", "Octal value specifying mode for attachment files written to disk. Must begin with '0' or '0o'.")
	dirCreateModeFlag     = flag.String("dirCreateMode", "0700", "Octal value specifying mode for attachment directories created on disk. Must begin with '0' or '0o'.")
)

func Main() error {
	ctx := context.Background()
	flag.Parse()

	if *configDirFlag != "" {
		_ = os.Setenv(envVarConfigDir, *configDirFlag)
	} else if os.Getenv(envVarConfigDir) == "" {
		flag.PrintDefaults()
		return fmt.Errorf("argument -configDir is required (if not using environment variable %s)", envVarConfigDir)
	}

	if *attachmentsDirFlag != "" {
		_ = os.Setenv(envVarAttachmentsDir, *attachmentsDirFlag)
	} else if os.Getenv(envVarAttachmentsDir) == "" {
		flag.PrintDefaults()
		return fmt.Errorf("argument -attachmentsDir is required (if not using environment variable %s)", envVarAttachmentsDir)
	}

	if *attachmentsDirURLFlag != "" {
		_ = os.Setenv(envVarAttachmentsDirURL, *attachmentsDirURLFlag)
	} else if os.Getenv(envVarAttachmentsDirURL) == "" {
		flag.PrintDefaults()
		return fmt.Errorf("argument -attachmentsDirURL is required (if not using environment variable %s)", envVarAttachmentsDirURL)
	}

	if *incomingEmailFlag != "" {
		_ = os.Setenv(envVarIncomingEmail, *incomingEmailFlag)
	} else if os.Getenv(envVarIncomingEmail) == "" {
		flag.PrintDefaults()
		return fmt.Errorf("argument 'incomingEmail' is required (if not using environment variable %s)", envVarIncomingEmail)
	}

	if *outgoingEmailFlag != "" {
		_ = os.Setenv(envVarOutgoingEmail, *outgoingEmailFlag)
	} else if os.Getenv(envVarOutgoingEmail) == "" {
		flag.PrintDefaults()
		return fmt.Errorf("argument 'outgoingEmail' is required (if not using environment variable %s)", envVarOutgoingEmail)
	}

	var fileCreateMode os.FileMode
	if mode, err := strconv.ParseInt(*fileCreateModeFlag, 8, 64); err != nil {
		flag.PrintDefaults()
		return errors.New("fileCreateMode must be an octal value parsable by strconv.ParseInt")
	} else {
		fileCreateMode = os.FileMode(mode)
	}

	var dirCreateMode os.FileMode
	if mode, err := strconv.ParseInt(*dirCreateModeFlag, 8, 64); err != nil {
		flag.PrintDefaults()
		return errors.New("dirCreateMode must be an octal value parsable by strconv.ParseInt")
	} else {
		dirCreateMode = os.FileMode(mode)
	}

	srv, err := buildGmailService(ctx)
	if err != nil {
		return err
	}

	var messagesToProcess []*gmail.Message
	searchQuery := fmt.Sprintf("to:\"%s\" is:unread", MustGetenv(envVarIncomingEmail))
	if err = srv.Users.Messages.List("me").IncludeSpamTrash(false).Q(searchQuery).Context(ctx).Pages(ctx, func(response *gmail.ListMessagesResponse) error {
		for _, mStub := range response.Messages {
			if m, err := srv.Users.Messages.Get("me", mStub.Id).Context(ctx).Do(); err != nil {
				return fmt.Errorf("error fetching message %s: %w", mStub.Id, err)
			} else {
				messagesToProcess = append(messagesToProcess, m)
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("error fetching messages to process: %w", err)
	}
	if len(messagesToProcess) == 0 {
		log.Println("no messages found that require processing")
		return nil
	} else {
		log.Printf("found %d messages to process", len(messagesToProcess))
	}

	mdConv := md.NewConverter("", true, &md.Options{
		LinkStyle: "referenced",
	})

	for _, m := range messagesToProcess {
		subject := MessageSubject(m)
		outgoingBody, cidMap, err := processPayload(ctx, srv, mdConv, m.Id, m.Payload, fileCreateMode, dirCreateMode)
		if err != nil {
			return err
		}

		for cid, url := range cidMap {
			outgoingBody = strings.ReplaceAll(outgoingBody, "cid:"+cid, url)
		}

		var outgoingMessage gmail.Message
		outgoingMessage.Raw = base64.URLEncoding.EncodeToString([]byte(
			"From: " + MustGetenv(envVarIncomingEmail) + "\r\n" +
				"To: " + MustGetenv(envVarOutgoingEmail) + "\r\n" +
				"Subject: " + subject + "\r\n\r\n" + outgoingBody))
		if _, err = srv.Users.Messages.Send("me", &outgoingMessage).Context(ctx).Do(); err != nil {
			return fmt.Errorf("failed to send message to Things (%s): %w", MustGetenv(envVarOutgoingEmail), err)
		}

		if !GetenvBool(envVarDontTouchOrigMessage, false) {
			if _, err := srv.Users.Messages.Modify("me", m.Id, &gmail.ModifyMessageRequest{RemoveLabelIds: []string{"UNREAD"}}).Context(ctx).Do(); err != nil {
				return fmt.Errorf("failed to mark message %s as read", m.Id)
			}
			if _, err := srv.Users.Messages.Trash("me", m.Id).Context(ctx).Do(); err != nil {
				return fmt.Errorf("failed to trash message %s", m.Id)
			}
		}

		log.Printf("processsed message %s (\"%s\")", m.Id, subject)
	}

	return nil
}

// processPayload returns the text representing this payload part, and a map of CID -> URL for any attachments processed in the part.
func processPayload(ctx context.Context, srv *gmail.Service, mdConv *md.Converter, messageId string, payload *gmail.MessagePart, fileCreateMode, dirCreateMode os.FileMode) (string, map[string]string, error) {
	if payload.MimeType == "text/plain" {
		bodyBytes, err := base64.URLEncoding.DecodeString(payload.Body.Data)
		if err != nil {
			return "", nil, err
		}
		return string(bodyBytes), nil, nil
	} else if payload.MimeType == "text/html" {
		bodyBytes, err := base64.URLEncoding.DecodeString(payload.Body.Data)
		if err != nil {
			return "", nil, err
		}
		parsed, err := mdConv.ConvertString(string(bodyBytes))
		if err != nil {
			return "", nil, err
		}
		return parsed + "\r\n\r\n", nil, nil
	} else if strings.HasPrefix(payload.MimeType, "multipart/") {
		outgoingBody := ""
		cidMap := make(map[string]string)
		for _, part := range payload.Parts {
			partBody, partCidMap, err := processPayload(ctx, srv, mdConv, messageId, part, fileCreateMode, dirCreateMode)
			if err != nil {
				return "", nil, err
			}
			outgoingBody = outgoingBody + partBody
			for k, v := range partCidMap {
				cidMap[k] = v
			}
		}
		return outgoingBody, cidMap, nil
	} else if payload.Body.AttachmentId != "" {
		attachmentUrl, cid, err := writeAttachmentFromPartReturningURLAndCID(ctx, srv, messageId, payload, fileCreateMode, dirCreateMode)
		if err != nil {
			return "", nil, err
		}
		return attachmentUrl + "\r\n\r\n", map[string]string{ cid: attachmentUrl }, nil
	} else {
		log.Printf("warning: could not parse message part %v", *payload)
		return "", nil, nil
	}
}

func writeAttachmentFromPartReturningURLAndCID(ctx context.Context, srv *gmail.Service, messageId string, part *gmail.MessagePart, fileCreateMode, dirCreateMode os.FileMode) (string, string, error) {
	dir, dirURL, err := attachmentsDirAndURL(messageId, dirCreateMode)
	if err != nil {
		return "", "", err
	}
	response, err := srv.Users.Messages.Attachments.Get("me", messageId, part.Body.AttachmentId).Context(ctx).Do()
	if err != nil {
		return "", "", fmt.Errorf("failed to download attachment %s for message %s: %w", part.Body.AttachmentId, messageId, err)
	}
	data, err := base64.URLEncoding.DecodeString(response.Data)
	if err != nil {
		return "", "", fmt.Errorf("failed to decode attachment %s for message %s: %w", part.Body.AttachmentId, messageId, err)
	}
	attachmentFilename := part.Filename
	if attachmentFilename == "" {
		attachmentFilename = messageId
		extCandidates, err := mime.ExtensionsByType(part.MimeType)
		if err != nil && extCandidates != nil && len(extCandidates) > 0 {
			attachmentFilename = attachmentFilename + extCandidates[0]
		}
	}
	fullFilePath := path.Join(dir, attachmentFilename) // full path to the attachment file on disk
	writtenAttachmentName := ""                        // name of the attachment file, as successfully written to disk
	i := 0
	for {
		fullPathToTryWriting := fullFilePath
		if i != 0 {
			ext := filepath.Ext(fullFilePath)
			fullPathToTryWriting = strings.TrimSuffix(fullFilePath, ext) + " (" + strconv.Itoa(i) + ")" + ext
		}
		err = WriteFileExcl(fullPathToTryWriting, data, fileCreateMode)
		if err != nil && os.IsExist(err) {
			i++
			continue
		} else if err != nil {
			return "", "", fmt.Errorf("failed to write attachment %s for message %s to path %s: %w", part.Body.AttachmentId, messageId, fullFilePath, err)
		} else {
			writtenAttachmentName = filepath.Base(fullPathToTryWriting)
			break
		}
	}
	return dirURL + "/" + url.PathEscape(writtenAttachmentName), PartCID(part), nil
}

func attachmentsDirAndURL(messageId string, dirCreateMode os.FileMode) (string, string, error) {
	dir := MustGetenv(envVarAttachmentsDir)
	dir = path.Join(dir, messageId)
	if err := os.MkdirAll(dir, dirCreateMode); err != nil {
		return "", "", fmt.Errorf("failed to make attachments dir %s: %w", dir, err)
	}
	dirURL := MustGetenv(envVarAttachmentsDirURL) + "/" + messageId
	return dir, dirURL, nil
}

func main() {
	if err := Main(); err != nil {
		log.Fatalf(err.Error())
	}
}
