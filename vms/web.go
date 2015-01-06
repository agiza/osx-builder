package vms

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	govix "github.com/hooklift/govix"

	"github.com/c4milo/go-osx-builder/apperror"
	"github.com/c4milo/go-osx-builder/config"
	"github.com/julienschmidt/httprouter"
	"github.com/satori/go.uuid"
)

func Init(router *httprouter.Router) {
	log.Infoln("Initializing vms module...")

	router.POST("/vms", CreateVM)
	router.GET("/vms", ListVMs)
	router.GET("/vms/:id", GetVM)
	router.DELETE("/vms/:id", DestroyVM)
}

type CreateVMParams struct {
	CPUs             uint              `json:"cpus"`
	Memory           string            `json:"memory"`
	NetType          govix.NetworkType `json:"network_type"`
	OSImage          Image             `json:"image"`
	BootstrapScript  string            `json:"bootstrap_script"`
	ToolsInitTimeout time.Duration     `json:"tools_init_timeout"`
	LaunchGUI        bool              `json:"launch_gui"`
	CallbackURL      string            `json:"callback_url"`
}

func sendResult(url string, obj interface{}) {
	if url == "" {
		return
	}
	data, err := json.Marshal(obj)
	if err != nil {
		log.WithFields(log.Fields{
			"vm":         obj,
			"code":       ErrCreatingVM.Code,
			"error":      err.Error(),
			"stacktrace": apperror.GetStacktrace(),
		}).Error(ErrCbURL.Message)

		data, err = json.Marshal(ErrCbURL)
		if err != nil {
			return
		}
	}
	_, err = http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.WithFields(log.Fields{
			"vm":         obj,
			"code":       ErrCbURL.Code,
			"error":      err.Error(),
			"stacktrace": apperror.GetStacktrace(),
		}).Error(ErrCbURL.Message)
	}
}

func CreateVM(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	r := config.Render

	var params CreateVMParams
	body, err := ioutil.ReadAll(req.Body)

	if err != nil {
		log.WithFields(log.Fields{
			"code":       ErrReadingReqBody.Code,
			"error":      err.Error(),
			"stacktrace": apperror.GetStacktrace(),
		}).Error(ErrReadingReqBody.Message)

		r.JSON(w, ErrReadingReqBody.HTTPStatus, ErrReadingReqBody)
		return
	}

	err = json.Unmarshal(body, &params)
	if err != nil {
		log.WithFields(log.Fields{
			"code":       ErrParsingJSON.Code,
			"error":      err.Error(),
			"stacktrace": apperror.GetStacktrace(),
		}).Error(ErrParsingJSON.Message)

		r.JSON(w, ErrParsingJSON.HTTPStatus, ErrParsingJSON)
		return
	}

	name := uuid.NewV4()

	vm := &VM{
		Provider:         govix.VMWARE_WORKSTATION,
		VerifySSL:        false,
		Name:             name.String(),
		Image:            params.OSImage,
		CPUs:             params.CPUs,
		Memory:           params.Memory,
		UpgradeVHardware: false,
		ToolsInitTimeout: params.ToolsInitTimeout,
		LaunchGUI:        params.LaunchGUI,
	}

	nic := &govix.NetworkAdapter{
		ConnType: params.NetType,
	}

	vm.VNetworkAdapters = make([]*govix.NetworkAdapter, 0, 1)
	vm.VNetworkAdapters = append(vm.VNetworkAdapters, nic)

	go func() {
		id, err := vm.Create()
		if err != nil {
			log.WithFields(log.Fields{
				"vm":         vm,
				"code":       ErrCreatingVM.Code,
				"error":      err.Error(),
				"stacktrace": apperror.GetStacktrace(),
			}).Error(ErrCreatingVM.Message)

			sendResult(params.CallbackURL, ErrCreatingVM)
			return
		}

		if vm.IPAddress == "" {
			vm.Refresh(id)
		}

		sendResult(params.CallbackURL, vm)
	}()

	r.JSON(w, http.StatusAccepted, vm)
}

type DestroyVMParams struct {
	ID string
}

func DestroyVM(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	r := config.Render

	params := DestroyVMParams{
		ID: ps.ByName("id"),
	}

	vm, err := FindVM(params.ID)
	if err != nil {
		log.WithFields(log.Fields{
			"code":       ErrOpeningVM.Code,
			"error":      err.Error(),
			"stacktrace": apperror.GetStacktrace(),
		}).Error(ErrOpeningVM.Message)

		r.JSON(w, ErrOpeningVM.HTTPStatus, ErrOpeningVM)
		return
	}

	if vm == nil {
		log.WithFields(log.Fields{
			"code":       ErrVMNotFound.Code,
			"error":      "",
			"stacktrace": "",
		}).Error(ErrVMNotFound.Message)

		r.JSON(w, ErrVMNotFound.HTTPStatus, ErrVMNotFound)
		return
	}

	err = vm.Destroy(vm.VMXFile)
	if err != nil {
		log.WithFields(log.Fields{
			"code":       ErrInternal.Code,
			"error":      err.Error(),
			"stacktrace": apperror.GetStacktrace(),
		}).Error(ErrInternal.Message)

		r.JSON(w, ErrInternal.HTTPStatus, ErrInternal)
		return
	}

	r.JSON(w, http.StatusNoContent, nil)
}

type ListVMsParams struct {
	Status string
}

func ListVMs(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	r := config.Render

	// params := ListVMsParams{
	// 	Status: req.URL.Query().Get("status"),
	// }

	vm := VM{
		Provider:  govix.VMWARE_WORKSTATION,
		VerifySSL: false,
	}

	host, err := vm.client()
	if err != nil {
		log.WithFields(log.Fields{
			"code":       ErrInternal.Code,
			"error":      err.Error(),
			"stacktrace": apperror.GetStacktrace(),
		}).Error(ErrInternal.Message)

		r.JSON(w, ErrInternal.HTTPStatus, ErrInternal)
		return
	}
	defer host.Disconnect()

	ids, err := host.FindItems(govix.FIND_RUNNING_VMS)
	if err != nil {
		log.WithFields(log.Fields{
			"code":       ErrInternal.Code,
			"error":      err.Error(),
			"stacktrace": apperror.GetStacktrace(),
		}).Error(ErrInternal.Message)

		r.JSON(w, ErrInternal.HTTPStatus, ErrInternal)
		return
	}

	r.JSON(w, http.StatusOK, ids)
}

type GetVMParams struct {
	ID string
}

func GetVM(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	r := config.Render

	params := GetVMParams{
		ID: ps.ByName("id"),
	}

	vm, err := FindVM(params.ID)

	if err != nil {
		log.WithFields(log.Fields{
			"code":       ErrOpeningVM.Code,
			"error":      err.Error(),
			"stacktrace": apperror.GetStacktrace(),
		}).Error(ErrOpeningVM.Message)

		r.JSON(w, ErrOpeningVM.HTTPStatus, ErrOpeningVM)
		return
	}

	if vm == nil {
		log.WithFields(log.Fields{
			"code":       ErrVMNotFound.Code,
			"error":      "",
			"stacktrace": "",
		}).Error(ErrVMNotFound.Message)

		r.JSON(w, ErrVMNotFound.HTTPStatus, ErrVMNotFound)
		return
	}

	r.JSON(w, http.StatusOK, vm)
}