# `mailto-things`: Pseudo-attachment support for emails to Things.app

## What It Does

**The problem:** The personal task manager application [Things](https://culturedcode.com/things/) does not support attaching files/images to tasks. This is one of my main complaints about the app compared to eg. [OmniFocus](https://www.omnigroup.com/omnifocus/).

**The solution:** `mailto-things` will:
1. Check for incoming emails in a Gmail account
2. Strip out the attachments from each email and put the files in a directory tree on disk
3. Replace the attachments in the email by URLs linking to the attachment files
4. Extract text from image attachments and include it in the email's body alongside attachment links
5. Send the modified email along to Things via [the app's "Mail to Things" feature](https://culturedcode.com/things/support/articles/2908262/)

The end result is that you can email files to your specially-configured email address, and the resulting tasks in your Things inbox contain links to those files. This is as close as I can get to actually attaching files directly to things.

## Installation & Setup

### macOS via Homebrew

```shell
brew install cdzombak/oss/mailto-things
```

### Debian via apt repository

Install my Debian repository if you haven't already:

```shell
sudo apt-get install ca-certificates curl gnupg
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://dist.cdzombak.net/deb.key | sudo gpg --dearmor -o /etc/apt/keyrings/dist-cdzombak-net.gpg
sudo chmod 0644 /etc/apt/keyrings/dist-cdzombak-net.gpg
echo -e "deb [signed-by=/etc/apt/keyrings/dist-cdzombak-net.gpg] https://dist.cdzombak.net/deb/oss any oss\n" | sudo tee -a /etc/apt/sources.list.d/dist-cdzombak-net.list > /dev/null
sudo apt-get update
```

Then install `mailto-things` via `apt-get`:

```shell
sudo apt-get install mailto-things
```

### Manual installation from build artifacts

Pre-built binaries for Linux and macOS on various architectures are downloadable from each [GitHub Release](https://github.com/cdzombak/mailto-things/releases). Debian packages for each release are available as well.

### Build and install locally

```shell
git clone https://github.com/cdzombak/mailto-things.git
cd mailto-things
make build

cp out/mailto-things $INSTALL_DIR
```

### Gmail App Credentials

You will need to create a Google Cloud Platform project with access to the Gmail API. The easiest way to do that is Step 1 on Google's [Gmail API Go Quickstart](https://developers.google.com/gmail/api/quickstart/go) documentation page. Download the resulting `credentials.json` file and store it in a private `mailto-things` configuration directory on the server you'll use to run this program.

### OCR Setup

The `-ocr` option will attempt to extract text from image attachments and include the text in the task's description (along with the link to the attachment). This requires a Google Cloud project set up, with billing and the Vision API enabled, and a local file with Google service account credentials.

- See [the image label detection guide](https://cloud.google.com/vision/docs/detect-labels-image-client-libraries) for details on Cloud Vision API setup
- See [the application default credentials docs](https://cloud.google.com/docs/authentication/application-default-credentials) for details on credentials setup
- Create a service account at [Service Accounts in the GCP Console](https://console.cloud.google.com/iam-admin/serviceaccounts). Only the `roles/storage.objectViewer` role is needed.
- Create and download a JSON key for the service account from the GCP Console.
- Give the path to the service account JSON key as the `GOOGLE_APPLICATION_CREDENTIALS` environment variable.

Note that using OCR is optional.

### Web Hosting

Attachment files must be written to a web-accessible directory. This is beyond the scope of this README, but basically you just need a writable folder which the web server can read & serve from. (I strongly recommend disabling directory listing on that folder!)

I place attachment files in a special directory on my home NAS server, and I make them accessible on my Tailscale network using [`tailscale serve`](https://tailscale.com/kb/1242/tailscale-serve/). I also use [Syncthing](https://syncthing.net) to sync the folder to my laptops; this means that, in addition to my attachment files being accessible anywhere on the Tailnet, they're synchronized locally for offline use.

## Usage

The first time you run the program you'll have to authorize its access to your Gmail account. Therefore, test your configuration by running `mailto-things` in an interactive shell before setting up a cronjob.

The following arguments (or their equivalent environment variables) are required:

- `-configDir`: Path to the directory where `credentials.json` & user authorization tokens are stored. (Overrides environment variable `MAILTO_THINGS_CONFIG_DIR`.)
- `-attachmentsDir`: Path to the directory on disk where attachments are stored. (Overrides environment variable `MAILTO_THINGS_ATTACHMENTS_DIR`.)
- `-attachmentsDirURL`: URL to the directory where attachments are stored. Should not end with a slash. (Overrides environment variable `MAILTO_THINGS_ATTACHMENTS_DIR_URL`.)
- `-incomingEmail`: Email address which receives tasks with attachments. This is your new "Things Inbox" email address, to which you will send messages. (Overrides environment variable `MAILTO_THINGS_INCOMING_EMAIL`.)
- `-outgoingEmail`: [Mail to Things](https://culturedcode.com/things/support/articles/2908262/) email address to send task emails to. (Overrides environment variable `MAILTO_THINGS_OUTGOING_EMAIL`.)

The following arguments are not strictly required, but you will almost definitely need to set them:

- `-fileCreateMode`: Octal value specifying the [permissions](https://web.archive.org/web/20201207170802/https://www.grymoire.com/Unix/Permissions.html) used to create file attachments on disk. Must begin with `0` or `0o`. Defaults to `0600`.
- `-dirCreateMode`: Octal value specifying the [permissions](https://web.archive.org/web/20201207170802/https://www.grymoire.com/Unix/Permissions.html) used to create attachment directories on disk. Must begin with `0` or `0o`. Defaults to `0700`.

The following arguments are not required:

- `-dontTouchOrigMessage`: If given, the original message will not be marked as read or trashed. (Overrides environment variable `MAILTO_THINGS_DONT_TOUCH_ORIG_MESSAGE`.) This is particularly useful when testing your setup.
- `-ocr` will attempt to extract text from image attachments and include it in the task's description. See "OCR," below.
- `-help` will print help and exit.
- `-version` will print the version number and exit.

### Cron Example

Here's an example of running `mailto-things` periodically via cron, adapted from my own usage:

`
*/5  *  *  *  *  runner -print-if-match "could not parse message part" -work-dir /home/cdzombak/mailto-things -- ./mailto-things -configDir /home/cdzombak/mailto-things/config -attachmentsDir /home/cdzombak/mailto-things/attachments -attachmentsDirURL "https://dropbox.dzombak.com/mailto-things" -fileCreateMode 0644 -dirCreateMode 0755 -incomingEmail "example+mailtothings@gmail.com" -outgoingEmail "add-to-things-example@things.email"
`

This example uses my [`runner` tool](https://github.com/cdzombak/runner) ([introductory blog post](https://www.dzombak.com/blog/2020/12/Introducing-Runner-a-lightweight-wrapper-for-cron-jobs.html)) to avoid emailing me output unless something went wrong.

## Docker

Docker images are available for a variety of Linux architectures from Docker Hub and GHCR. Images are based on the `scratch` image and are as small as possible.

> TODO(cdzombak): Links.

Run them via, for example:

```shell
docker run --rm \
    -v /home/cdzombak/mailto-things/config:/app-config \
    -v /home/cdzombak/mailto-things/attachments:/app-attachments \
    -e GOOGLE_APPLICATION_CREDENTIALS=/app-config/google-sa-credentials.json \
    cdzombak/mailto-things:1 \
    -configDir /app-config \
    -attachmentsDir /app-attachments \
    -attachmentsDirURL "https://example.tailnet-1234.ts.net" \
    -fileCreateMode 0644 \
    -dirCreateMode 0755 \
    -incomingEmail "example+mailtothings@gmail.com" \
    -outgoingEmail "add-to-things-example@things.email"
```

Keep in mind that all paths given as arguments are paths _within the container,_ including the `GOOGLE_APPLICATION_CREDENTIALS` environment variable, so you'll have to make sure they are all mapped to the desired paths on the host.

## License

GNU LGPL v3; see LICENSE in this repository.

## About

- Issues: [github.com/cdzombak/mailto-things/issues](https://github.com/cdzombak/mailto-things/issues)
- Author: [Chris Dzombak](https://www.dzombak.com) ([GitHub @cdzombak](https://github.com/cdzombak))
