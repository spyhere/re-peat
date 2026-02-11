package editor

type tags struct {
	mLife         *struct{}
	soundWave     *struct{}
	noneArea      *struct{}
	mCreateButton *struct{}
	backdrop      *struct{}
}

func newTags() *tags {
	return &tags{
		mLife:         &struct{}{},
		soundWave:     &struct{}{},
		noneArea:      &struct{}{},
		mCreateButton: &struct{}{},
		backdrop:      &struct{}{},
	}
}
