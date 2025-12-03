// Package renamer provides functionality for renaming media files based on metadata
// from external providers.
//
// The package supports renaming movies, TV shows, seasons, and episodes by:
//   - Scanning source paths for media files
//   - Fetching metadata from configured providers
//   - Formatting new filenames using customizable formatters
//   - Creating organized directory structures
//   - Supporting multiple rename modes (symlink, copy)
//
// The renamer handles duplicate detection, path deduplication, and provides
// detailed logging of all operations. It supports dry-run mode for previewing
// changes before applying them.
package renamer
