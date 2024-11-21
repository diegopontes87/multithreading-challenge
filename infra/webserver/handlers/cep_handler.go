package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/diegopontes87/multithreading-challenge/internal/entity"
	"github.com/go-chi/chi"
)

var (
	cepBrasilURL = "https://brasilapi.com.br/api/cep/v1/%s"
	viaCEPURL    = "https://viacep.com.br/ws/%s/json/"
)

func GetCEPInfo(w http.ResponseWriter, r *http.Request) {

	channel1 := make(chan []byte)
	channel2 := make(chan []byte)
	cep := chi.URLParam(r, "cep")
	go getViaCEPInfo(cep, w, channel1)
	go getBrasilCEPInfo(cep, w, channel2)
	selectResponse(w, channel1, channel2)

}

func selectResponse(w http.ResponseWriter, channel1 chan []byte, channel2 chan []byte) {
	select {
	case respViaCep := <-channel1:
		var viacep entity.VIACepAddress
		err := json.Unmarshal(respViaCep, &viacep)
		if err != nil {
			log.Printf("Error during parse: %v:", err)
		}
		log.Printf("Request made in: %s portal\n", viaCEPURL)
		log.Println("Response from ViaCep\n", &viacep)

		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&viacep)
	case respBrasilCep := <-channel2:
		var brasilAPICEP entity.BrasilAPIAddress
		err := json.Unmarshal(respBrasilCep, &brasilAPICEP)
		if err != nil {
			log.Printf("Error during parse: %v:", err)
		}
		log.Printf("Request made in: %s portal\n", cepBrasilURL)
		log.Println("Response from CepBrasil:\n", &brasilAPICEP)

		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&brasilAPICEP)
	case <-time.After(1 * time.Second):
		http.Error(w, "Request timed out", http.StatusRequestTimeout)

	}
}

func getBrasilCEPInfo(cep string, w http.ResponseWriter, ch chan []byte) {

	url := fmt.Sprintf(cepBrasilURL, cep)
	resp, err := http.Get(url)

	if handleRequestError(err, "BrasilAPI") {
		ch <- nil
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading BrasilAPIAddress response: %v:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ch <- body
}

func getViaCEPInfo(cep string, w http.ResponseWriter, ch chan []byte) {

	url := fmt.Sprintf(viaCEPURL, cep)
	resp, err := http.Get(url)

	if handleRequestError(err, "ViaCEP") {
		ch <- nil
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading ViaCEP response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ch <- body
}

func handleRequestError(err error, apiName string) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		log.Printf("%s request timed out: %v", apiName, err)
	} else {
		log.Printf("Error during %s request: %v", apiName, err)
	}
	return true
}
