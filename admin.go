package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/GlazeLab/PureGamer/src/model"
	"github.com/GlazeLab/PureGamer/src/modules/entry"
	"github.com/GlazeLab/PureGamer/src/modules/exit"
	"github.com/GlazeLab/PureGamer/src/modules/optimizer"
	"github.com/GlazeLab/PureGamer/src/modules/pinging"
	"github.com/GlazeLab/PureGamer/src/modules/relaying"
	"github.com/GlazeLab/PureGamer/src/modules/superadmin"
	"github.com/GlazeLab/PureGamer/src/node"
	"github.com/GlazeLab/PureGamer/src/utils"
	logging "github.com/ipfs/go-log/v2"
	"log"
	"net/http"
)

func main() {
	ctx := context.TODO()
	n, err := node.Listen(ctx)
	if err != nil {
		panic(err)
	}

	admin, err := superadmin.NewSuperAdmin(n)
	if err != nil {
		panic(err)
	}
	exits, err := exit.NewExit(n)
	if err != nil {
		panic(err)
	}
	err = relaying.Register(n, exits)
	if err != nil {
		panic(err)
	}
	err = pinging.Register(n)
	if err != nil {
		panic(err)
	}
	optimized, err := optimizer.NewOptimizer(n)
	if err != nil {
		panic(err)
	}

	admin.Handle(ctx)
	optimized.Handle(ctx)
	go optimized.RunSpeedTest(ctx)

	err = entry.Listen(n, exits, optimized)
	if err != nil {
		panic(err)
	}

	logging.SetAllLoggers(logging.LevelInfo)

	privateKeyText, err := utils.ReadText("private.key")
	privKey, err := utils.DecodePrivate(privateKeyText)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		// read put
		if r.Method == "PUT" {
			var config model.Config
			err := json.NewDecoder(r.Body).Decode(&config)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			err = admin.SendConfig(ctx, config, privKey)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
			return
		} else if r.Method == "GET" {
			// return current config
			err := json.NewEncoder(w).Encode(n.Config)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
	})
	http.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		graphText := optimized.Info()
		w.WriteHeader(http.StatusOK)
		body := fmt.Sprintf(`<body>
Network Topology:
  <pre class="mermaid">
	graph LR
	%s
  </pre>
  <script type="module">
    import mermaid from 'https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.esm.min.mjs';
    mermaid.initialize({ startOnLoad: true });
  </script>
</body>`, graphText)
		w.Write([]byte(body))
		return
	})
	http.HandleFunc("/info/route", func(w http.ResponseWriter, r *http.Request) {
		from := r.URL.Query().Get("from")
		to := r.URL.Query().Get("to")
		graphText := optimized.RouteInfo(from, to)
		w.WriteHeader(http.StatusOK)
		body := fmt.Sprintf(`<body>
Route from %s to %s:
  <pre class="mermaid">
	graph LR
	%s
  </pre>
  <script type="module">
    import mermaid from 'https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.esm.min.mjs';
	mermaid.initialize({ startOnLoad: true });
  </script>
</body>`, from, to, graphText)
		w.Write([]byte(body))
		return
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
