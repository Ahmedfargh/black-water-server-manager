package background

type BackgroundTask interface {
	Run() (int32, error)
	HandleError(err error) error
}
