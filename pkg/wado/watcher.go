package wado

// Watcher watches a directories and files for changes and posts them to
// the channel gotten with GetChannel(). The string posted is the path of the changed file.
type Watcher interface {
	FileCount() int
	CreateChangeChannel() chan string
	Close() error
	AddCallback(func(string))
}
