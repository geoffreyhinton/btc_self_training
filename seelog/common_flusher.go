package seelog

// flusherInterface represents all objects that have to do cleanup
// at certain moments of time (e.g. before app shutdown to avoid data loss)
type flusherInterface interface {
	Flush()
}
