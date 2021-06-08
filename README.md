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

## Installation

To install the binary on your server, clone this repository and run `make install`. If you prefer to build the program elsewhere, the Makefile provides some targets for cross-compilation, eg. `make build-linux-amd64`. Run `make help` for a list of targets. The build products are placed in `./out`; just copy the `mailto-things` binary to wherever you want to run it.

### Web Hosting

Attachment files must be written to a web-accessible directory. This is beyond the scope of this README, but basically you just need a writable folder which the web server can read & serve from. I strongly recommend disabling directory listing on that folder!

I'm placing attachment files in my [public "Dropbox" folder](https://www.dzombak.com/blog/2014/01/serving-dropbox-via-nginx.html), though I now use [Syncthing](https://syncthing.net) instead of Dropbox for sharing stuff between my computers & servers. This means that, in addition to my attachment files being web-accessible, they're synchronized locally to my computers for offline use.

### Gmail App Credentials

You will need to create a Google Cloud Platform project with access to the Gmail API. The easiest way to do that is Step 1 on Google's [Gmail API Go Quickstart](https://developers.google.com/gmail/api/quickstart/go) documentation page. Download the resulting `credentials.json` file and store it in a private `mailto-things` configuration directory on the server you'll use to run this program.

## Usage & Configuration

Note that the first time you run the program you'll have to authorize its access to your Gmail account. Therefore, test your configuration by running `mailto-things` in an interactive shell before setting up a cronjob.

All of the following arguments (or their equivalent environment variables) are required:

- `-configDir`: Path to the directory where `credentials.json` & user authorization tokens are stored. (Overrides environment variable `MAILTO_THINGS_CONFIG_DIR`.)
- `-attachmentsDir`: Path to the directory on disk where attachments are stored. (Overrides environment variable `MAILTO_THINGS_ATTACHMENTS_DIR`.)
- `-attachmentsDirURL`: URL to the directory where attachments are stored. Should not end with a slash. (Overrides environment variable `MAILTO_THINGS_ATTACHMENTS_DIR_URL`.)
- `-incomingEmail`: Email address which receives tasks with attachments. This is your new "Things Inbox" email address, to which you will send messages. (Overrides environment variable `MAILTO_THINGS_INCOMING_EMAIL`.)
- `-outgoingEmail`: [Mail to Things](https://culturedcode.com/things/support/articles/2908262/) email address to send task emails to. (Overrides environment variable `MAILTO_THINGS_OUTGOING_EMAIL`.)

The following arguments are not strictly required, but you will almost definitely need to set them:

- `-fileCreateMode`: Octal value specifying the [permissions](https://web.archive.org/web/20201207170802/https://www.grymoire.com/Unix/Permissions.html) used to create file attachments on disk. Must begin with `0` or `0o`. Defaults to `0600`.
- `-dirCreateMode`: Octal value specifying the [permissions](https://web.archive.org/web/20201207170802/https://www.grymoire.com/Unix/Permissions.html) used to create attachment directories on disk. Must begin with `0` or `0o`. Defaults to `0700`.

The following arguments are not required:

- `-ocr` will attempt to extract text from image attachments and include it in the task's description. See "OCR," below.
- `-version` will print the version number and exit.

### Cron Example

Here's an example of running `mailto-things` periodically via cron, adapted from my own usage:

`
*/5		*	*	*	*	runner -print-if-match "could not parse message part" -work-dir /home/cdzombak/mailto-things -- ./mailto-things -configDir /home/cdzombak/mailto-things -attachmentsDir /home/cdzombak/Sync/public/mailto-things -attachmentsDirURL "https://dropbox.dzombak.com/mailto-things" -fileCreateMode 0644 -dirCreateMode 0755 -incomingEmail "example+mailtothings@gmail.com" -outgoingEmail "add-to-things-example@things.email"
`

This example uses my [`runner` tool](https://github.com/cdzombak/runner) ([introductory blog post](https://www.dzombak.com/blog/2020/12/Introducing-Runner--a-lightweight-wrapper-for-cron-jobs.html)) to avoid emailing me output unless something went wrong.

## OCR

The `-ocr` option will attempt to extract text from image attachments and include the text in the task's description, along with the link to the attachment. This requires the command-line tools `tesseract` and `ispell` to be available in the PATH.

On Ubuntu, these packages are available via `apt install tesseract ispell`.

On macOS, these packages are available via Homebrew: `brew install tesseract ispell`.

## About

- Issues: https://github.com/cdzombak/mailto-things/issues/new
- Author: [Chris Dzombak](https://www.dzombak.com)
