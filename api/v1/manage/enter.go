package manage

import "main.go/service"

type ManageGroup struct {
	ManageAdminUserApi
	ManageAlertApi
}

var finopsAdminUserService = service.ServiceGroupApp.ManageServiceGroup.ManageAdminUserService
var finopsAdminUserTokenService = service.ServiceGroupApp.ManageServiceGroup.ManageAdminUserTokenService
var fileUploadAndDownloadService = service.ServiceGroupApp.ExampleServiceGroup.FileUploadAndDownloadService
var finopsAlertService = service.ServiceGroupApp.ManageServiceGroup.ManageAlertService
