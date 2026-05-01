//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -lobjc

#import <Cocoa/Cocoa.h>
#import <objc/runtime.h>
#include <stdlib.h>

// Ring buffer for file open events
#define MAX_OPEN_FILES 32
static char* openFileQueue[MAX_OPEN_FILES];
static int openFileWriteIdx = 0;
static int openFileReadIdx = 0;

static void enqueueFile(const char* path) {
    @synchronized([NSApp class]) {
        int idx = openFileWriteIdx % MAX_OPEN_FILES;
        if (openFileQueue[idx] != NULL) {
            free(openFileQueue[idx]);
            openFileQueue[idx] = NULL;
        }
        openFileQueue[idx] = strdup(path);
        openFileWriteIdx++;
    }
}

// Method to inject into Fyne's delegate
static BOOL injected_openFile(id self, SEL _cmd, NSApplication* sender, NSString* filename) {
    enqueueFile([filename UTF8String]);
    return YES;
}

static void injected_openFiles(id self, SEL _cmd, NSApplication* sender, NSArray* filenames) {
    for (NSString* filename in filenames) {
        enqueueFile([filename UTF8String]);
    }
    [sender replyToOpenOrPrint:NSApplicationDelegateReplySuccess];
}

static void injectDelegateMethods() {
    id delegate = [[NSApplication sharedApplication] delegate];
    if (delegate == nil) return;

    Class cls = [delegate class];

    // Use class_replaceMethod to override Fyne's default implementation
    // (class_addMethod silently fails if the method already exists)
    class_replaceMethod(cls, @selector(application:openFile:),
        (IMP)injected_openFile, "B@:@@");
    class_replaceMethod(cls, @selector(application:openFiles:),
        (IMP)injected_openFiles, "v@:@@");
}

static void installOpenFileSupport() {
    // Register a notification observer that fires BEFORE macOS delivers the
    // application:openFile: Apple Event during app launch.  This avoids the
    // race where dispatch_after(0.1s) runs too late and the default Fyne
    // delegate rejects the file with "cannot open files in the 'text' format".
    [[NSNotificationCenter defaultCenter]
        addObserverForName:NSApplicationWillFinishLaunchingNotification
        object:nil
        queue:nil
        usingBlock:^(NSNotification* note) {
            injectDelegateMethods();
        }];

    // Fallback: also inject after a short delay in case the notification path
    // didn't fire (e.g. app was already running when a new file is opened).
    dispatch_after(dispatch_time(DISPATCH_TIME_NOW, (int64_t)(0.1 * NSEC_PER_SEC)), dispatch_get_main_queue(), ^{
        injectDelegateMethods();
    });
}

static const char* pollOpenFile() {
    const char* result = NULL;
    @synchronized([NSApp class]) {
        if (openFileReadIdx < openFileWriteIdx) {
            int idx = openFileReadIdx % MAX_OPEN_FILES;
            result = openFileQueue[idx];
            openFileQueue[idx] = NULL;
            openFileReadIdx++;
        }
    }
    return result;
}
*/
import "C"

import (
	"sync"
	"time"
	"unsafe"
)

var (
	openFileChan = make(chan string, 10)
	openFileOnce sync.Once
)

func initOpenFileHandler() <-chan string {
	openFileOnce.Do(func() {
		C.installOpenFileSupport()
		go func() {
			for {
				cpath := C.pollOpenFile()
				if cpath != nil {
					path := C.GoString(cpath)
					C.free(unsafe.Pointer(cpath))
					openFileChan <- path
				} else {
					time.Sleep(200 * time.Millisecond)
				}
			}
		}()
	})
	return openFileChan
}

