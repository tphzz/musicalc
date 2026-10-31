//go:build darwin && !ios && cgo

// NSOpenPanel and NSSavePanel for native_darwin.go. The C strings are copied into NSStrings
// before returning, because Go frees them right after the call; the panels then run on the main
// queue and report the chosen path (or NULL when cancelled) to musicalcPanelDone.

#import <AppKit/AppKit.h>
#import <UniformTypeIdentifiers/UniformTypeIdentifiers.h>
#include <stdint.h>

extern void musicalcPanelDone(int handle, char* path);

// musicalcAllowExtension limits a panel to files with the extension
static void musicalcAllowExtension(NSSavePanel* panel, NSString* ext) {
	if (ext.length == 0) {
		return;
	}
	if (@available(macOS 11.0, *)) {
		UTType* type = [UTType typeWithFilenameExtension:ext];
		if (type != nil) {
			panel.allowedContentTypes = @[type];
			return;
		}
	}
#pragma clang diagnostic push
#pragma clang diagnostic ignored "-Wdeprecated-declarations"
	panel.allowedFileTypes = @[ext];
#pragma clang diagnostic pop
}

// musicalcRunPanel shows the panel as a sheet on the window, or on its own without one
static void musicalcRunPanel(NSSavePanel* panel, uintptr_t window, int handle) {
	void (^done)(NSModalResponse) = ^(NSModalResponse result) {
		if (result == NSModalResponseOK && panel.URL != nil) {
			musicalcPanelDone(handle, (char*)panel.URL.path.fileSystemRepresentation);
		} else {
			musicalcPanelDone(handle, NULL);
		}
	};
	NSWindow* win = (NSWindow*)(void*)window;
	if (win != nil) {
		[panel beginSheetModalForWindow:win completionHandler:done];
	} else {
		[panel beginWithCompletionHandler:done];
	}
}

void musicalcOpenPanel(uintptr_t window, const char* title, const char* ext, int handle) {
	NSString* t = [NSString stringWithUTF8String:title];
	NSString* e = [NSString stringWithUTF8String:ext];
	dispatch_async(dispatch_get_main_queue(), ^{
		NSOpenPanel* panel = [NSOpenPanel openPanel];
		panel.message = t;
		panel.canChooseFiles = YES;
		panel.canChooseDirectories = NO;
		panel.allowsMultipleSelection = NO;
		musicalcAllowExtension(panel, e);
		musicalcRunPanel(panel, window, handle);
	});
}

void musicalcSavePanel(uintptr_t window, const char* title, const char* name, const char* ext, int handle) {
	NSString* t = [NSString stringWithUTF8String:title];
	NSString* n = [NSString stringWithUTF8String:name];
	NSString* e = [NSString stringWithUTF8String:ext];
	dispatch_async(dispatch_get_main_queue(), ^{
		NSSavePanel* panel = [NSSavePanel savePanel];
		panel.message = t;
		panel.nameFieldStringValue = n;
		panel.canCreateDirectories = YES;
		musicalcAllowExtension(panel, e);
		musicalcRunPanel(panel, window, handle);
	});
}
