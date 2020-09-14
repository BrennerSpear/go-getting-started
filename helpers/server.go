package helpers

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

"github.com/go-chi/chi"
"github.com/go-chi/chi/middleware"
"github.com/go-chi/cors"
"github.com/go-chi/render"
"github.com/pires/go-proxyproto"
log "github.com/sirupsen/logrus"

)

type Result struct {
	ProcessedString string         `json:"processedString"`
	RawISPInfo      IPInfoResponse `json:"rawIspInfo"`
}

const (
	// chunk size is 1 mib
	chunkSize = 1048576
)

var (
	// generate random data for download test on start to minimize runtime overhead
	randomData = getRandomData(chunkSize)
)

func ListenAndServe(conf *Config) error {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.GetHead)

	cs := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS", "HEAD"},
		AllowedHeaders: []string{"*"},
	})

	r.Use(cs.Handler)
	r.Use(middleware.NoCache)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)


	conf.Port = os.Getenv("PORT")
	addr := net.JoinHostPort(conf.BindAddress, conf.Port)
	log.Infof("Starting backend server on localhost%s", addr)
	//r.Get("/*", pages)
	r.HandleFunc("/empty", empty)
	r.HandleFunc("/backend/empty", empty)
	r.Get("/garbage", garbage)
	r.Get("/backend/garbage", garbage)
	r.Get("/getIP", getIP)
	r.Get("/backend/getIP", getIP)

	go listenProxyProtocol(conf, r)
	return http.ListenAndServe(addr, r)
}

func listenProxyProtocol(conf *Config, r *chi.Mux) {
	if conf.ProxyProtocolPort != "0" {
		addr := net.JoinHostPort(conf.BindAddress, conf.ProxyProtocolPort)
		l, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatal("Cannot listen on proxy protocol port %s: %s", conf.ProxyProtocolPort, err)
		}

		pl := &proxyproto.Listener{Listener: l}
		defer pl.Close()

		log.Infof("Starting proxy protocol listener on %s", addr)
		log.Fatal(http.Serve(pl, r))
	}
}

//func pages(w http.ResponseWriter, r *http.Request) {
//	if r.RequestURI == "/" {
//		r.RequestURI = "/index.html"
//	}
//
//	conf := config.LoadedConfig()
//	fs := justFilesFilesystem{fs: http.Dir(conf.AssetsPath), readDirBatchSize: 2}
//	http.FileServer(fs).ServeHTTP(w, r)
//}

func empty(w http.ResponseWriter, r *http.Request) {
	io.Copy(ioutil.Discard, r.Body)
	r.Body.Close()

	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
}

func garbage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Description", "File Transfer")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=random.dat")
	w.Header().Set("Content-Transfer-Encoding", "binary")

	// chunk size set to 4 by default
	chunks := 4

	ckSize := r.FormValue("ckSize")
	if ckSize != "" {
		i, err := strconv.ParseInt(ckSize, 10, 64)
		if err != nil {
			log.Errorf("Invalid chunk size: %s", ckSize)
			log.Warn("Will use default value %d", chunks)
		} else {
			// limit max chunk size to 1024
			if i > 1024 {
				chunks = 1024
			} else {
				chunks = int(i)
			}
		}
	}

	for i := 0; i < chunks; i++ {
		if _, err := w.Write(randomData); err != nil {
			log.Errorf("Error writing back to client at chunk number %d: %s", i, err)
			break
		}
	}
}

func getIP(w http.ResponseWriter, r *http.Request) {
	var ret Result

	clientIP := r.RemoteAddr
	clientIP = strings.ReplaceAll(clientIP, "::ffff:", "")

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		clientIP = ip
	}

	isSpecialIP := true
	switch {
	case clientIP == "::1":
		ret.ProcessedString = clientIP + " - localhost IPv6 access"
	case strings.HasPrefix(clientIP, "fe80:"):
		ret.ProcessedString = clientIP + " - link-local IPv6 access"
	case strings.HasPrefix(clientIP, "127."):
		ret.ProcessedString = clientIP + " - localhost IPv4 access"
	case strings.HasPrefix(clientIP, "10."):
		ret.ProcessedString = clientIP + " - private IPv4 access"
	case regexp.MustCompile(`^172\.(1[6-9]|2\d|3[01])\.`).MatchString(clientIP):
		ret.ProcessedString = clientIP + " - private IPv4 access"
	case strings.HasPrefix(clientIP, "192.168"):
		ret.ProcessedString = clientIP + " - private IPv4 access"
	case strings.HasPrefix(clientIP, "169.254"):
		ret.ProcessedString = clientIP + " - link-local IPv4 access"
	case regexp.MustCompile(`^100\.([6-9][0-9]|1[0-2][0-7])\.`).MatchString(clientIP):
		ret.ProcessedString = clientIP + " - CGNAT IPv4 access"
	default:
		isSpecialIP = false
	}

	if isSpecialIP {
		b, _ := json.Marshal(&ret)
		if _, err := w.Write(b); err != nil {
			log.Errorf("Error writing to client: %s", err)
		}
		return
	}

	getISPInfo := r.FormValue("isp") == "true"
	distanceUnit := r.FormValue("distance")

	ret.ProcessedString = clientIP

	if getISPInfo {
		ispInfo := getIPInfo(clientIP)
		ret.RawISPInfo = ispInfo

		removeRegexp := regexp.MustCompile(`AS\d+\s`)
		isp := removeRegexp.ReplaceAllString(ispInfo.Organization, "")

		if isp == "" {
			isp = "Unknown ISP"
		}

		if ispInfo.Country != "" {
			isp += ", " + ispInfo.Country
		}

		if ispInfo.Location != "" {
			isp += " (" + calculateDistance(ispInfo.Location, distanceUnit) + ")"
		}

		ret.ProcessedString += " - " + isp
	}

	render.JSON(w, r, ret)
}
