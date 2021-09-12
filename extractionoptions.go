package squashfs

import "io/fs"

//ExtractionOptions holds options that are available when extracting Files.
//When using, start with DefaultOptions() in case new options are added.
type ExtractionOptions struct {
	DereferenceSymlink bool        //Instead of extracting symlinks as symlinks, extract the referenced file in it's place.
	UnbreakSymlink     bool        //When extracting a symlink, tries to extract the referenced file along with it (if pointed to within the squashfs)
	Verbose            bool        //Print info about what's being extracted. Prints to the log library.
	AllowErrors        bool        //Ignore errors when they occur. If Verbose is set, prints the errors.
	FolderPerm         fs.FileMode //The permission the extraction folder is set to.
}

//DefaultOptions is the default extraction options.
//These are the options used for ExtractTo.
func DefaultOptions() ExtractionOptions {
	return ExtractionOptions{
		DereferenceSymlink: false,
		UnbreakSymlink:     false,
		Verbose:            false,
		AllowErrors:        false,
		FolderPerm:         fs.ModePerm,
	}
}
