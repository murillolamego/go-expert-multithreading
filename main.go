package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type ApiResponse struct {
	val []byte
	api string
	err error
}

type BrasilAPI struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	Service      string `json:"service"`
}

type ViaCep struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", buscaCepHandler)
	http.ListenAndServe(":8080", mux)
}

func buscaCepHandler(w http.ResponseWriter, r *http.Request) {
	apiTimeoutMilli := 1000
	apiCtx, apiCancel := context.WithTimeout(context.Background(), time.Duration(apiTimeoutMilli)*time.Millisecond)
	defer apiCancel()
	respApiCh := make(chan ApiResponse)

	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	cepParam := r.URL.Query().Get("cep")
	if cepParam == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	go func() {
		val, api, err := buscaCepBrasilApi(apiCtx, cepParam)
		respApiCh <- ApiResponse{
			val: val,
			api: api,
			err: err,
		}
	}()

	go func() {
		val, api, err := buscaCepViaCep(apiCtx, cepParam)
		respApiCh <- ApiResponse{
			val: val,
			api: api,
			err: err,
		}
	}()

	for {
		select {
		case <-apiCtx.Done():
			println("Err: Fetching data from external API took too long")
			w.WriteHeader(http.StatusRequestTimeout)
			return
		case apiResp := <-respApiCh:
			apiCancel()
			if apiResp.err != nil {
				println("Err: Internal Server Error")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			println(string(apiResp.val))
			println("Via: " + apiResp.api)

			if len(apiResp.val) == 0 {
				println("Err: Empty body")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			switch apiResp.api {
			case "BrasilAPI":
				var cepInfo BrasilAPI
				err := json.Unmarshal(apiResp.val, &cepInfo)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(cepInfo)
				return
			case "ViaCep":
				var cepInfo ViaCep
				err := json.Unmarshal(apiResp.val, &cepInfo)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(cepInfo)
				return
			}
		}
	}

}

func buscaCepBrasilApi(ctx context.Context, cep string) ([]byte, string, error) {
	println("Api Call - BrasilAPI")
	req, err := http.NewRequestWithContext(ctx, "GET", "https://brasilapi.com.br/api/cep/v1/"+cep, nil)
	if err != nil {
		return nil, "BrasilAPI", err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "BrasilAPI", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, "BrasilAPI", err
	}

	return body, "BrasilAPI", nil
}

func buscaCepViaCep(ctx context.Context, cep string) ([]byte, string, error) {
	println("Api Call - ViaCEP")
	req, err := http.NewRequestWithContext(ctx, "GET", "http://viacep.com.br/ws/"+cep+"/json/", nil)
	if err != nil {
		return nil, "ViaCep", err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "ViaCep", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, "ViaCep", err
	}

	return body, "ViaCep", nil
}
