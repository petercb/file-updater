# Bugfix Requirements Document

## Introduction

When file-updater runs as a sidecar container and the YAML config is updated (triggering a re-fetch via fsnotify), the `Download` function fails to overwrite existing files at the destination. This means updated versions of remote files are never written to disk, defeating the core purpose of the tool in watch mode.

The root cause is that `go-getter`'s `Client` with `Mode: getter.ClientModeAny` does not overwrite existing files/directories at the destination path. The fix must ensure that stale destination paths are removed before each download so that `go-getter` always writes fresh content.

## Bug Analysis

### Current Behavior (Defect)

1.1 WHEN a file has already been downloaded to a destination AND the config triggers a re-fetch (via fsnotify or initial fetch of an already-existing path) THEN the system fails to overwrite the existing file with the newer version from the remote source

1.2 WHEN a directory has already been downloaded to a destination AND the config triggers a re-fetch THEN the system fails to overwrite the existing directory with the newer version from the remote source

### Expected Behavior (Correct)

2.1 WHEN a file has already been downloaded to a destination AND the config triggers a re-fetch THEN the system SHALL remove the existing destination file and download the new version, resulting in the destination containing the updated content

2.2 WHEN a directory has already been downloaded to a destination AND the config triggers a re-fetch THEN the system SHALL remove the existing destination directory and download the new version, resulting in the destination containing the updated content

### Unchanged Behavior (Regression Prevention)

3.1 WHEN the destination path does not already exist THEN the system SHALL CONTINUE TO download the file/directory to the destination successfully (first-time download)

3.2 WHEN the source URL is invalid or unreachable THEN the system SHALL CONTINUE TO return an appropriate error without modifying the destination

3.3 WHEN the URL detection step fails THEN the system SHALL CONTINUE TO return a detection error without modifying the filesystem

3.4 WHEN the destination path does not exist and the parent directory is writable THEN the system SHALL CONTINUE TO create the destination file/directory as before
