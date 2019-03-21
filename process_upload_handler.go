package uadmin

import "net/http"

// custom upload
type ProcessUploadInterface interface {
	ProcessUpload(r *http.Request, f *F, modelName string, session *Session, s *ModelSchema) (val string, ok bool)
}

func DoProcessUpload(r *http.Request, f *F, modelName string, session *Session, s *ModelSchema) (val string) {
	if DefaultProcessUploader != nil {
		val, ok := DefaultProcessUploader.ProcessUpload(r, f, modelName, session, s)
		if ok {
			return val
		}
	}
	return processUpload(r, f, modelName, session, s)
}
