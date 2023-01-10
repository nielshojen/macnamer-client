# Macnamer Client

The macOS client-side companion for [Macnamer](https://github.com/nielshojen/macnamer)

## Building it

This requires [The Luggage](https://github.com/unixorn/luggage) and that you have developer certificates handy.

The package uses a built in python and for that and the installer to be signed, you need to build it yourself.

Edit these two lines in the Makefile to set your developer certs:

```PB_EXTRA_ARGS+= --sign "Developer ID Installer: [COMPANY NAME] ([IDENTIFIER])"
DEV_APP_CERT="Developer ID Application: [COMPANY NAME] ([IDENTIFIER])"
```

and run:

```make pkg
```

## Deploying to clients

Installation is simple, install the package using your favourite method, and then set the preference to point to the ServerURL. You can use MCX, Profiles, or plain old defaults write:

```defaults write /Library/Preferences/com.nielshojen.macnamer ServerURL "https://macnamer.yourserver.com"
defaults write /Library/Preferences/com.nielshojen.macnamer Key "[ComputerGroup Key]"
```

## Acknowledgements

This project was originally started by Graham Gilbert, so all kudos go to him. This project is just to keep this alive as long as i need it, and the change of pref domain is just to keep the two separate.
