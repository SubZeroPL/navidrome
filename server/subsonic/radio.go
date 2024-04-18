package subsonic

import (
	"bufio"
	"io"
	"net/http"
	"net/url"

	"github.com/navidrome/navidrome/log"
	"github.com/navidrome/navidrome/model"
	"github.com/navidrome/navidrome/server/subsonic/responses"
	"github.com/navidrome/navidrome/utils/req"
)

func (api *Router) CreateInternetRadio(r *http.Request) (*responses.Subsonic, error) {
	p := req.Params(r)
	streamUrl, err := p.String("streamUrl")
	if err != nil {
		return nil, err
	}

	name, err := p.String("name")
	if err != nil {
		return nil, err
	}

	homepageUrl, _ := p.String("homepageUrl")
	ctx := r.Context()

	radio := &model.Radio{
		StreamUrl:   streamUrl,
		HomePageUrl: homepageUrl,
		Name:        name,
	}

	err = api.ds.Radio(ctx).Put(radio)
	if err != nil {
		return nil, err
	}
	return newResponse(), nil
}

func (api *Router) DeleteInternetRadio(r *http.Request) (*responses.Subsonic, error) {
	p := req.Params(r)
	id, err := p.String("id")

	if err != nil {
		return nil, err
	}

	err = api.ds.Radio(r.Context()).Delete(id)
	if err != nil {
		return nil, err
	}
	return newResponse(), nil
}

func (api *Router) GetInternetRadios(r *http.Request) (*responses.Subsonic, error) {
	ctx := r.Context()
	radios, err := api.ds.Radio(ctx).GetAll(model.QueryOptions{Sort: "name"})
	if err != nil {
		return nil, err
	}

	res := make([]responses.Radio, len(radios))
	for i, g := range radios {
		res[i] = responses.Radio{
			ID:          g.ID,
			Name:        g.Name,
			StreamUrl:   g.StreamUrl,
			HomepageUrl: g.HomePageUrl,
		}
	}

	response := newResponse()
	response.InternetRadioStations = &responses.InternetRadioStations{
		Radios: res,
	}

	return response, nil
}

func (api *Router) UpdateInternetRadio(r *http.Request) (*responses.Subsonic, error) {
	p := req.Params(r)
	id, err := p.String("id")
	if err != nil {
		return nil, err
	}

	streamUrl, err := p.String("streamUrl")
	if err != nil {
		return nil, err
	}

	name, err := p.String("name")
	if err != nil {
		return nil, err
	}

	homepageUrl, _ := p.String("homepageUrl")
	ctx := r.Context()

	radio := &model.Radio{
		ID:          id,
		StreamUrl:   streamUrl,
		HomePageUrl: homepageUrl,
		Name:        name,
	}

	err = api.ds.Radio(ctx).Put(radio)
	if err != nil {
		return nil, err
	}
	return newResponse(), nil
}

// The following endpoints are not part of the subsonic spec, but are part of the
// Subsonic router (as opposed to native) because it makes authentication easier

func (api *Router) proxyIcon(w http.ResponseWriter, r *http.Request) {
	iconUrl, err := requiredParamString(r, "url")

	if err != nil {
		log.Error(r, "Bad stream url", "url", iconUrl, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	client := http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", iconUrl, nil)
	if err != nil {
		log.Error(r, "Error creating request", "url", iconUrl, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Error(r, "Error fetching icon", "url", iconUrl, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)

	if err != nil {
		log.Error(r, "Error fetching icon", "url", iconUrl, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

var (
	headers = []string{"ice-audio-info", "icy-br", "icy-description", "icy-genre", "icy-name", "icy-pub", "icy-sr", "icy-url", "icy-vbr", "icy-metaint"}
)

func (api *Router) proxyRadio(w http.ResponseWriter, r *http.Request) {
	requestedUrl, err := requiredParamString(r, "url")

	if err != nil {
		log.Error(r, "Bad stream url", "url", requestedUrl, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	client := http.Client{}

	streamUrl, err := url.QueryUnescape(requestedUrl)

	if err != nil {
		log.Error(r, "Bad stream url", "url", requestedUrl, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req, err := http.NewRequestWithContext(ctx, "HEAD", streamUrl, nil)

	if err != nil {
		log.Error(r, "Error creating request", "url", streamUrl, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Set("Icy-Metadata", "1")
	headResp, err := client.Do(req)

	if err != nil {
		log.Error(r, "Error fetching stream", "url", streamUrl, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	headResp.Body.Close()

	req, _ = http.NewRequestWithContext(ctx, "GET", streamUrl, nil)
	req.Header.Set("Icy-Metadata", "1")
	mainResp, err := client.Do(req)

	if err != nil {
		log.Error(r, "Error fetching stream", "url", streamUrl, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, header := range headers {
		val := headResp.Header.Get(header)
		if val != "" {
			w.Header().Set(header, val)
		}
	}

	w.Header().Set("Content-Type", mainResp.Header.Get("Content-Type"))

	defer mainResp.Body.Close()
	reader := bufio.NewReader(mainResp.Body)
	buf := make([]byte, 8192)

	done := false

	go func() {
		<-r.Context().Done()
		done = true
	}()

	for {
		count, err := reader.Read(buf)

		if count == 0 || done {
			break
		}

		if err != nil {
			log.Error(r, "Error reading data", "url", streamUrl, err)
			break
		}

		_, err = w.Write(buf[0:count])

		if err != nil {
			log.Error(r, "Error writing data", "url", streamUrl, err)
			break
		}
	}
}
