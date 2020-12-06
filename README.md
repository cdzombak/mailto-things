# `mailto-things`: Pseudo-attachment support for emails to Things.app

## What It Does

TK: what problem does this solve, and how?

## Installation

TODO(cdzombak): make install target

TK: install docs

## Usage

TK: how do you run this?

### Cron Example

`
*/5		*	*	*	*	runner -print-if-match "could not parse message part" -work-dir /home/cdzombak/mailto-things -- ./mailto-things -configDir /home/cdzombak/mailto-things -attachmentsDir /home/cdzombak/Sync/public/mailto-things -attachmentsDirURL "https://dropbox.dzombak.com/mailto-things" -fileCreateMode 0644 -dirCreateMode 0755 -incomingEmail "example+things@gmail.com" -outgoingEmail "add-to-things-example@things.email"
`

TK: link to runner / blog post

## About

- Issues: https://github.com/cdzombak/mailto-things/issues/new
- Author: [Chris Dzombak](https://www.dzombak.com)
