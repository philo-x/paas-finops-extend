package manage

import "main.go/service"

type ManageGroup struct {
	ManageAdminUserApi
}

var adminUserService = service.ServiceGroupApp.ManageServiceGroup.ManageAdminUserService
var adminUserTokenService = service.ServiceGroupApp.ManageServiceGroup.ManageAdminUserTokenService
var fileUploadAndDownloadService = service.ServiceGroupApp.ExampleServiceGroup.FileUploadAndDownloadService
