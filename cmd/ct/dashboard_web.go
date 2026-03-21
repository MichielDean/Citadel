package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// newDashboardMux returns an http.Handler for the web dashboard.
// Exposed for testing.
func newDashboardMux(cfgPath, dbPath string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, dashboardHTML)
	})

	mux.HandleFunc("/api/dashboard", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		data := fetchDashboardData(cfgPath, dbPath)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data) //nolint:errcheck
	})

	mux.HandleFunc("/api/dashboard/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		send := func() {
			data := fetchDashboardData(cfgPath, dbPath)
			b, err := json.Marshal(data)
			if err != nil {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", b)
			flusher.Flush()
		}

		send()

		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case <-ticker.C:
				send()
			}
		}
	})

	return mux
}

// RunDashboardWeb starts the HTTP web dashboard on addr and blocks until
// SIGINT/SIGTERM is received or the server fails.
func RunDashboardWeb(cfgPath, dbPath, addr string) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           newDashboardMux(cfgPath, dbPath),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      0, // SSE streams are long-lived
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Fprintf(os.Stderr, "Cistern web dashboard listening on http://localhost%s\n", addr)

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutCtx)
	case err := <-errCh:
		return err
	}
}

// dashboardHTML is the single-page web dashboard.
const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Cistern</title>
<style>
:root {
  --bg:#0d1117; --surface:#161b22; --border:#30363d;
  --text:#e6edf3; --dim:#7d8590; --green:#3fb950;
  --yellow:#d29922; --red:#f85149; --blue:#58a6ff;
  --font:'Courier New',Courier,monospace;
}
@media (prefers-color-scheme:light) {
  :root {
    --bg:#ffffff; --surface:#f6f8fa; --border:#d0d7de;
    --text:#1f2328; --dim:#636c76; --green:#1a7f37;
    --yellow:#9a6700; --red:#cf222e; --blue:#0969da;
  }
}
*{box-sizing:border-box;margin:0;padding:0}
body{background:var(--bg);color:var(--text);font-family:var(--font);font-size:13px;line-height:1.5;padding:12px;min-height:100vh}
h1{font-size:15px;color:var(--dim);letter-spacing:2px;text-transform:uppercase;margin-bottom:10px}
.sep{border:none;border-top:1px solid var(--border);margin:10px 0}
.section-title{font-size:10px;color:var(--dim);letter-spacing:1px;text-transform:uppercase;margin-bottom:6px}
#conn{font-size:11px;margin-bottom:8px}
.live{color:var(--green)}
.offline{color:var(--red)}
/* Aqueduct cards */
.aqueduct{border:1px solid var(--border);border-radius:4px;margin-bottom:8px;overflow:hidden}
.aq-name{font-size:10px;color:var(--dim);padding:3px 8px;background:var(--surface);border-bottom:1px solid var(--border)}
.aq-channel{padding:5px 8px;text-align:center;font-size:12px;border-bottom:1px solid var(--border)}
.aq-channel.active{color:var(--green);background:rgba(63,185,80,.08)}
.aq-channel.idle{color:var(--dim)}
.aq-piers{display:flex;align-items:flex-start;padding:8px 4px;overflow-x:auto}
.aq-pier{display:flex;flex-direction:column;align-items:center;min-width:64px;flex:1}
.pier-connector{flex:1;height:2px;background:var(--border);min-width:4px;margin-top:12px;align-self:flex-start}
.pier-dot{width:26px;height:26px;border-radius:50%;border:2px solid var(--dim);display:flex;align-items:center;justify-content:center;font-size:11px;color:var(--dim)}
.pier-dot.active{border-color:var(--green);color:var(--green);background:rgba(63,185,80,.1)}
.pier-name{font-size:10px;color:var(--dim);text-align:center;margin-top:3px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;max-width:64px;padding:0 2px}
.pier-name.active{color:var(--green)}
/* Stats */
.stats{display:flex;gap:20px;flex-wrap:wrap;margin:8px 0}
.stat-num{font-size:22px;font-weight:bold}
.stat-lbl{font-size:10px;color:var(--dim);text-transform:uppercase}
.flowing .stat-num{color:var(--green)}
.queued .stat-num{color:var(--yellow)}
.done .stat-num{color:var(--dim)}
/* Flow activities */
.activity{border:1px solid var(--border);border-radius:3px;padding:8px;margin-bottom:6px}
.act-header{display:flex;gap:8px;align-items:baseline;flex-wrap:wrap;margin-bottom:3px}
.act-id{color:var(--blue);font-size:12px;white-space:nowrap}
.act-title{font-size:12px;flex:1;min-width:0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.act-step{font-size:11px;color:var(--green)}
.note{border-top:1px solid var(--border);margin-top:4px;padding-top:3px;font-size:11px}
.note-who{color:var(--blue)}
.note-body{color:var(--dim)}
/* Tables */
.items{width:100%;border-collapse:collapse}
.items td{padding:4px 4px;font-size:12px;border-bottom:1px solid var(--border);vertical-align:top}
.items tr:last-child td{border-bottom:none}
.t-id{color:var(--blue);white-space:nowrap}
.t-title{word-break:break-word}
.t-blocked{color:var(--red);font-size:11px;white-space:nowrap}
.s-delivered{color:var(--green)}
.s-stagnant{color:var(--red)}
.s-in_progress{color:var(--yellow)}
.empty{color:var(--dim);font-size:12px;padding:2px 0}
footer{color:var(--dim);font-size:11px;margin-top:12px;padding-top:8px;border-top:1px solid var(--border)}
@media(max-width:480px){body{padding:8px}}
</style>
</head>
<body>
<h1>&#x2697; Cistern</h1>
<div id="conn" class="offline">&#x25CB; connecting&#x2026;</div>
<div id="app"></div>
<script>
var app=document.getElementById('app');
var connEl=document.getElementById('conn');

function esc(s){
  if(s==null)return'';
  return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}

function bar(idx,total,w){
  if(total<=0||idx<=0)return'\u2591'.repeat(w);
  var f=Math.min(Math.floor(idx*w/total),w);
  return'\u2588'.repeat(f)+'\u2591'.repeat(w-f);
}

function fmtNs(ns){
  var s=Math.floor(ns/1e9);
  if(s<0)return'0s';
  if(s<60)return s+'s';
  var m=Math.floor(s/60);
  return m+'m '+(s%60)+'s';
}

function render(d){
  var h='';

  // Aqueducts
  h+='<div class="section"><div class="section-title">Aqueducts</div>';
  var chs=d.cataractae||[];
  if(!chs.length){
    h+='<div class="empty">No aqueducts configured.</div>';
  } else {
    for(var i=0;i<chs.length;i++){
      var ch=chs[i];
      h+='<div class="aqueduct">';
      h+='<div class="aq-name">'+esc(ch.name)+'</div>';
      if(ch.droplet_id){
        h+='<div class="aq-channel active">\u2248\u2248  '+esc(ch.droplet_id)+'  '+bar(ch.cataractae_index,ch.total_cataractae,8)+'  '+fmtNs(ch.elapsed)+'  \u2248\u2248</div>';
      } else {
        h+='<div class="aq-channel idle">\u2014 idle \u2014</div>';
      }
      var steps=ch.steps||[];
      if(steps.length){
        h+='<div class="aq-piers">';
        for(var j=0;j<steps.length;j++){
          var step=steps[j];
          var act=(step===ch.step&&!!ch.droplet_id);
          if(j>0)h+='<div class="pier-connector"></div>';
          h+='<div class="aq-pier">';
          h+='<div class="pier-dot'+(act?' active':'')+'">'+( act?'\u25cf':'\u25cb')+'</div>';
          h+='<div class="pier-name'+(act?' active':'')+'">'+esc(step)+'</div>';
          h+='</div>';
        }
        h+='</div>';
      }
      h+='</div>';
    }
  }
  h+='</div><hr class="sep">';

  // Stats
  h+='<div class="stats"><div class="stat flowing"><div class="stat-num">'+(d.flowing_count||0)+'</div><div class="stat-lbl">flowing</div></div>';
  h+='<div class="stat queued"><div class="stat-num">'+(d.queued_count||0)+'</div><div class="stat-lbl">queued</div></div>';
  h+='<div class="stat done"><div class="stat-num">'+(d.done_count||0)+'</div><div class="stat-lbl">delivered</div></div></div><hr class="sep">';

  // Current flow
  h+='<div class="section"><div class="section-title">Current Flow</div>';
  var acts=d.flow_activities||[];
  if(!acts.length){
    h+='<div class="empty">No active flow.</div>';
  } else {
    for(var i=0;i<acts.length;i++){
      var a=acts[i];
      h+='<div class="activity"><div class="act-header"><span class="act-id">'+esc(a.droplet_id)+'</span><span class="act-title">'+esc(a.title)+'</span></div>';
      h+='<div class="act-step">\u2192 '+esc(a.step)+'</div>';
      var notes=a.recent_notes||[];
      for(var k=0;k<notes.length;k++){
        var n=notes[k];
        h+='<div class="note"><span class="note-who">'+esc(n.cataractae_name)+'</span> <span class="note-body">'+esc(n.content)+'</span></div>';
      }
      h+='</div>';
    }
  }
  h+='</div><hr class="sep">';

  // Cistern
  h+='<div class="section"><div class="section-title">Cistern</div>';
  var queued=(d.cistern_items||[]).filter(function(x){return x.status==='open';});
  if(!queued.length){
    h+='<div class="empty">Cistern is empty.</div>';
  } else {
    h+='<table class="items">';
    for(var i=0;i<queued.length;i++){
      var item=queued[i];
      var blocked=d.blocked_by_map&&d.blocked_by_map[item.id];
      h+='<tr><td class="t-id">'+esc(item.id)+'</td><td class="t-title">'+esc(item.title)+'</td>';
      h+=blocked?'<td class="t-blocked">[blocked by '+esc(blocked)+']</td>':'<td></td>';
      h+='</tr>';
    }
    h+='</table>';
  }
  h+='</div><hr class="sep">';

  // Recent flow
  h+='<div class="section"><div class="section-title">Recent Flow</div>';
  var recent=d.recent_items||[];
  if(!recent.length){
    h+='<div class="empty">No recent flow.</div>';
  } else {
    h+='<table class="items">';
    for(var i=0;i<recent.length;i++){
      var item=recent[i];
      var t=item.updated_at?new Date(item.updated_at).toLocaleTimeString([],{hour:'2-digit',minute:'2-digit'}):'';
      h+='<tr><td class="t-id">'+esc(item.id)+'</td><td class="t-title">'+esc(item.title)+'</td>';
      h+='<td class="s-'+esc(item.status)+'">'+esc(item.status)+'</td><td style="color:var(--dim)">'+t+'</td></tr>';
    }
    h+='</table>';
  }
  h+='</div>';

  // Footer
  var ts=d.fetched_at?new Date(d.fetched_at).toLocaleTimeString():'';
  h+='<footer>last update: '+ts+'</footer>';

  app.innerHTML=h;
}

function connect(){
  connEl.className='offline';
  connEl.textContent='\u25cb connecting\u2026';
  var es=new EventSource('/api/dashboard/events');
  es.onopen=function(){
    connEl.className='live';
    connEl.textContent='\u25cf live';
  };
  es.onmessage=function(e){
    try{render(JSON.parse(e.data));}catch(err){console.error('cistern parse:',err);}
  };
  es.onerror=function(){
    connEl.className='offline';
    connEl.textContent='\u25cb reconnecting\u2026';
    es.close();
    setTimeout(connect,3000);
  };
}
connect();
</script>
</body>
</html>`
