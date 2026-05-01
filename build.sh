#!/bin/bash
set -e

APP_NAME="LogView"
ICON="Icon.png"

# Ensure fyne CLI is available
if ! command -v fyne &> /dev/null; then
    echo "Installing fyne CLI..."
    go install fyne.io/tools/cmd/fyne@latest
fi

# Add Go bin to PATH
export PATH="$PATH:$(go env GOPATH)/bin"

OS=$(uname -s)

case "$OS" in
    Darwin)
        echo "Building macOS .app bundle..."
        fyne package -os darwin -icon "$ICON" -name "$APP_NAME"
        # Add document type declarations so macOS allows opening files with this app
        PLIST="${APP_NAME}.app/Contents/Info.plist"
        # Remove existing entry first so re-builds don't silently fail
        /usr/libexec/PlistBuddy -c "Delete :CFBundleDocumentTypes" "$PLIST" 2>/dev/null || true
        /usr/libexec/PlistBuddy -c "Add :CFBundleDocumentTypes array" "$PLIST"
        /usr/libexec/PlistBuddy -c "Add :CFBundleDocumentTypes:0 dict" "$PLIST"
        /usr/libexec/PlistBuddy -c "Add :CFBundleDocumentTypes:0:CFBundleTypeName string 'All Files'" "$PLIST"
        /usr/libexec/PlistBuddy -c "Add :CFBundleDocumentTypes:0:CFBundleTypeRole string 'Viewer'" "$PLIST"
        /usr/libexec/PlistBuddy -c "Add :CFBundleDocumentTypes:0:LSHandlerRank string 'Alternate'" "$PLIST"
        /usr/libexec/PlistBuddy -c "Add :CFBundleDocumentTypes:0:LSItemContentTypes array" "$PLIST"
        /usr/libexec/PlistBuddy -c "Add :CFBundleDocumentTypes:0:LSItemContentTypes:0 string 'public.data'" "$PLIST"
        /usr/libexec/PlistBuddy -c "Add :CFBundleDocumentTypes:0:LSItemContentTypes:1 string 'public.content'" "$PLIST"
        /usr/libexec/PlistBuddy -c "Add :CFBundleDocumentTypes:0:LSItemContentTypes:2 string 'public.text'" "$PLIST"
        /usr/libexec/PlistBuddy -c "Add :CFBundleDocumentTypes:0:LSItemContentTypes:3 string 'public.plain-text'" "$PLIST"
        /usr/libexec/PlistBuddy -c "Add :CFBundleDocumentTypes:0:LSItemContentTypes:4 string 'public.log'" "$PLIST"
        /usr/libexec/PlistBuddy -c "Add :CFBundleDocumentTypes:0:LSItemContentTypes:5 string 'public.item'" "$PLIST"
        # Also declare extensions for broader compatibility
        /usr/libexec/PlistBuddy -c "Add :CFBundleDocumentTypes:0:CFBundleTypeExtensions array" "$PLIST"
        /usr/libexec/PlistBuddy -c "Add :CFBundleDocumentTypes:0:CFBundleTypeExtensions:0 string 'log'" "$PLIST"
        /usr/libexec/PlistBuddy -c "Add :CFBundleDocumentTypes:0:CFBundleTypeExtensions:1 string 'txt'" "$PLIST"
        /usr/libexec/PlistBuddy -c "Add :CFBundleDocumentTypes:0:CFBundleTypeExtensions:2 string '*'" "$PLIST"

        # Copy to Applications and reset Launch Services
        if [ -d "/Applications/${APP_NAME}.app" ]; then
            rm -rf "/Applications/${APP_NAME}.app"
        fi
        cp -R "${APP_NAME}.app" /Applications/
        /System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister -f "/Applications/${APP_NAME}.app" 2>/dev/null || true
        echo "Done: ${APP_NAME}.app (also copied to /Applications)"
        ;;
    Linux)
        echo "Building Linux package..."
        fyne package -os linux -icon "$ICON" -name "$APP_NAME"
        echo "Done: ${APP_NAME}.tar.xz"
        ;;
    MINGW*|MSYS*|CYGWIN*|Windows_NT)
        echo "Building Windows package..."
        fyne package -os windows -icon "$ICON" -name "$APP_NAME"
        echo "Done: ${APP_NAME}.exe"
        ;;
    *)
        echo "Unknown OS: $OS"
        exit 1
        ;;
esac
