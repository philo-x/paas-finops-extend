package webhook

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"main.go/service"
)

type RecommendationApi struct{}

var recommendationService = &service.ServiceGroupApp.WebhookServiceGroup.RecommendationService

func (r *RecommendationApi) ServeMutate(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, "could not read request")
		return
	}

	var admissionReviewReq admissionv1.AdmissionReview
	if err := json.Unmarshal(body, &admissionReviewReq); err != nil {
		c.JSON(http.StatusBadRequest, "could not unmarshal request")
		return
	}

	req := admissionReviewReq.Request
	var pod corev1.Pod
	if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
		c.JSON(http.StatusBadRequest, "could not unmarshal pod")
		return
	}

	patches, _ := recommendationService.MutatePod(&pod, req.Namespace)

	admissionResponse := &admissionv1.AdmissionResponse{
		UID:     req.UID,
		Allowed: true,
	}

	if len(patches) > 0 {
		patchBytes, _ := json.Marshal(patches)
		admissionResponse.Patch = patchBytes
		patchType := admissionv1.PatchTypeJSONPatch
		admissionResponse.PatchType = &patchType
	}

	admissionReviewResp := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{APIVersion: "admission.k8s.io/v1", Kind: "AdmissionReview"},
		Response: admissionResponse,
	}

	c.JSON(http.StatusOK, admissionReviewResp)
}
