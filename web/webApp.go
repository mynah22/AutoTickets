package web

import (
	"autotaskViewer/api"
	"autotaskViewer/secrets"
	"autotaskViewer/tickets"
	"fmt"
	"html/template"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type WebApp struct {
	E            *echo.Echo
	Sc           secrets.SecretsCollection
	Tc           tickets.TicketCollection
	wsClients    wsClients
	serverParams serverParams
}

type serverParams struct {
	sync.RWMutex
	pollRate     int
	port         int
	verboseApi   bool
	apiStartHour int
	apiEndHour   int
}

// returns pointer to a properly initialized WebApp value
func NewWebApp(
	logHttp bool,
	pollRate int,
	port int,
	saveFilePath string,
	verboseApi bool,
	apiStart int,
	apiEnd int) (w *WebApp) {

	ticketsSlice := make([]tickets.AutotaskTicket, 0)

	w = &WebApp{
		E:         echo.New(),
		Sc:        secrets.SecretsCollection{FilePath: saveFilePath},
		Tc:        tickets.TicketCollection{Tickets: &ticketsSlice},
		wsClients: wsClients{clients: make(map[*websocket.Conn]bool)},
		serverParams: serverParams{
			apiStartHour: apiStart,
			apiEndHour:   apiEnd,
			verboseApi:   verboseApi,
			pollRate:     pollRate,
			port:         port,
		},
	}

	if logHttp {
		w.E.Use(middleware.Logger())
	}
	w.E.Use(middleware.Recover())
	w.E.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root: "static",
	}))

	w.E.Renderer = &Template{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}

	w.E.GET("/", w.handleRoot)
	w.E.GET("/secrets", w.handleSecrets)
	w.E.POST("/submitSecrets", w.handleReceiveSecrets)
	// w.E.GET("/rscIdCount", func(c echo.Context) error {
	// 	w.RLock()
	// 	defer w.RUnlock()
	// 	return c.JSON(http.StatusOK, w.getRescIdCount())
	// })
	w.E.GET("/wsTickets", w.handleWsTickets)
	return w
}

// Starts serving clients and periodically polling API / updating websock clients
func (w *WebApp) Start() {
	go w.periodicallyPollApi()
	portStr := ":" + strconv.Itoa(w.serverParams.port)
	if err := w.E.Start(portStr); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

// periodically poll API and handle resulting data: update stored tickets and broadcast to websocket clients.
func (w *WebApp) periodicallyPollApi() {
	ticker := time.NewTicker(time.Duration(w.serverParams.pollRate) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		currentHour := time.Now().Hour()
		if currentHour < w.serverParams.apiStartHour || currentHour >= w.serverParams.apiEndHour {
			if w.serverParams.verboseApi {
				timeStamp := time.Now().Format("15:04 Jan 2")
				fmt.Printf(
					"\n[%v] API not queried (out of active hours)\n    (24hr times) API Start hour: %v, API End hour:%v\n",
					timeStamp, w.serverParams.apiStartHour, w.serverParams.apiEndHour,
				)
			}
		} else {
			if err := w.pollApi(); err != nil {
				w.E.Logger.Error("error polling api:", err)
			}
		}
	}
}

// obtain and handle API data
func (w *WebApp) pollApi() error {
	if !w.Sc.SecretsAreLoaded() {
		fmt.Println("Secrets not loaded, cannot poll API")
		return fmt.Errorf("secrets not loaded, cannot poll API")
	}
	freshTickets, err := api.GetOpenTickets(w.Sc.GetSecrets())

	if err != nil {
		fmt.Println("Error fetching tickets:", err)
		return err
	}
	if w.serverParams.verboseApi {
		timeStamp := time.Now().Format("15:04 Jan 2")
		fmt.Printf("\n[%v] fresh tickets obtained. Fresh open ticket count: %v", timeStamp, len(freshTickets))
	}
	w.Tc.SetTickets(&freshTickets)
	if w.serverParams.verboseApi {
		printHash := ""
		currentHash := w.Tc.GetCurrentHash()
		if len(currentHash) > 8 {
			printHash = string([]rune(currentHash)[:8])
		}
		fmt.Printf("\n  checking for new tickets hash. Old is '%v...'", printHash)
	}

	if w.Tc.CheckForNewHash() {
		if w.serverParams.verboseApi {
			currentHash := w.Tc.GetCurrentHash()
			fmt.Printf("\n  New tickets hash, sending broadcast. New hash is '%v...'\n", string([]rune(currentHash)[:8]))
		}
		go w.broadcastTickets()
	}

	return nil
}

//  helper types

// used to handle secrets submitted by user
type submittedSecrets struct {
	Username        string `json:"username"`
	IntegrationCode string `json:"integrationCode"`
	Secret          string `json:"secret"`
	Password        string `json:"password"`
}

// used for http templating
type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// func (w *WebApp) getRescIdCount() map[string]int {
// 	rescIdCount := make(map[string]int)
// 	for _, ticket := range *w.Tc {
// 		if ticket.AssignedResourceID != "" {
// 			rescIdCount[ticket.AssignedResourceID]++
// 		}
// 	}
// 	return rescIdCount
// }
