package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type ViaCep struct {
	Cep        string `json:"cep"`
	Logradouro string `json:"logradouro"`
	Bairro     string `json:"bairro"`
	Localidade string `json:"localidade"`
	Uf         string `json:"uf"`
}

type ApiCep struct {
	Cep        string `json:"code"`
	Logradouro string `json:"address"`
	Bairro     string `json:"district"`
	Localidade string `json:"city"`
	Uf         string `json:"state"`
}

type Error struct {
	Message string `json:"message"`
}

const urlViaCep = "https://viacep.com.br/ws/%s/json"
const urlApiCep = "https://cdn.apicep.com/file/apicep/%s.json"

func main() {
	http.HandleFunc("/", BuscarCepHandler)
	http.ListenAndServe(":8000", nil)
}

func BuscarCepHandler(w http.ResponseWriter, r *http.Request) {

	cep := r.URL.Query().Get("cep")

	if cep == "" {
		e := Error{Message: "cep is invalid"}
		json.NewEncoder(w).Encode(e)
		return
	}

	viaCep := make(chan interface{})
	apiCep := make(chan interface{})

	go BuscarViaCep(cep, viaCep)
	go BuscarApiCep(cep, apiCep)

	select {
	case v := <-viaCep:
		fmt.Print("Reposta Via Cep: ", v)
	case v := <-apiCep:
		fmt.Print("Reposta Api Cep: ", v)
	case <-time.After(1 * time.Second):
		json.NewEncoder(w).Encode(Error{Message: "exceeded timeout of 1 second :("})
	}
}

func BuscarViaCep(cep string, ch chan<- interface{}) {

	url := fmt.Sprintf(urlViaCep, cep)
	res, err := http.Get(url)
	if err != nil {
		er := Error{Message: err.Error()}
		ch <- er
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		er := Error{Message: err.Error()}
		ch <- er
		return
	}

	var viaCep ViaCep
	if err := json.Unmarshal(body, &viaCep); err != nil {
		er := Error{Message: err.Error()}
		ch <- er
		return
	}

	ch <- viaCep
}

func BuscarApiCep(cep string, ch chan<- interface{}) {
	url := fmt.Sprintf(urlApiCep, cep)
	res, err := http.Get(url)
	if err != nil {
		er := Error{Message: err.Error()}
		ch <- er
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		er := Error{Message: err.Error()}
		ch <- er
		return
	}
	var apiCep ApiCep
	if err := json.Unmarshal(body, &apiCep); err != nil {
		er := Error{Message: err.Error()}
		ch <- er
		return
	}

	ch <- apiCep
}
