package handler

import (
	"encoding/json"
	"fmt"
	"github.com/DmitryKhitrin/alerting-service/internal/common"
	"github.com/DmitryKhitrin/alerting-service/internal/server/metrics"
	"github.com/go-chi/chi"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Handler struct {
	service metrics.Service
}

func NewHandler(service metrics.Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) UpdatePlain(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")

	metricObj := &common.Metrics{}

	err := metricObj.CreateMetric(name, mType, value)

	if err != nil {
		http.Error(w, err.Text, err.Status)
		return
	}

	err = h.service.StoreMetric(metricObj)

	if err != nil {
		http.Error(w, err.Text, err.Status)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) UpdateJSON(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var metric common.Metrics
	err = json.Unmarshal(b, &metric)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	serviceErr := h.service.StoreMetric(&metric)

	if serviceErr != nil {
		http.Error(w, serviceErr.Text, serviceErr.Status)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetPlain(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")

	metricObj := &common.Metrics{}
	metricObj.FromNameAndType(name, mType)

	val, sErr := h.service.GetMetric(metricObj)

	if sErr != nil {
		http.Error(w, sErr.Text, sErr.Status)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(w, fmt.Sprint(val))
}

func (h *Handler) GetJSON(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var metric common.Metrics
	err = json.Unmarshal(b, &metric)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	_, sErr := h.service.GetMetric(&metric)

	if sErr != nil {
		log.Println(metric)
		log.Println(&err)
		http.Error(w, sErr.Text, sErr.Status)
		return
	}

	data, err := json.Marshal(&metric)
	if err != nil {
		log.Println("parsing error", err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	_, err = w.Write([]byte(data))
	if err != nil {
		return
	}
}

func (h *Handler) GetAllHandler(w http.ResponseWriter, _ *http.Request) {
	writeTmp, err := h.service.GetTemplateWriter()

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	err = writeTmp(w)
	if err != nil {
		log.Println(err)
		return
	}
}
