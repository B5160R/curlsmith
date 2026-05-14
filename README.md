# Curlsmith

                                                        
      ,- _~.             ,,                     ,  ,,    
    (' /|               ||                 '  ||  ||    
    ((  ||   \\ \\ ,._-_ ||  _-_, \\/\\/\\ \\ =||= ||/\\ 
    ((  ||   || ||  ||   || ||_.  || || || ||  ||  || || 
    ( / |   || ||  ||   ||  ~ || || || || ||  ||  || || 
      -____- \\/\\  \\,  \\ ,-_-  \\ \\ \\ \\  \\, \\ |/ 
                                                    _/  
                                                                     
                                     #+                               
                                #  #+++                               
                              #+   #+#+  #                            
                              #+#+#+#+#+#+#                           
                              +#+#+# #+#+#-                           
                            +  -#+##+  +#+                            
                            #++#  ++#+# #+  #                         
                            +##+####  + # #+#                         
                            #+#+# +#   #+#+#+                         
                             #+#      ## ##+                          
                               #+       +#                            
                                                                      
                            #++#+#+#+#+##+#+#+#+#+#+#+#+#+#+#+#       
       #+#+#+#+#+#+#+#+##+#+ #+##+#+#+#+++#+##+#+#+#+#+#+#+#+         
        #+#+#+++#+#+#+#+#+## #+#+#+#+++##+##+#+##+#+#+##              
          ##+#+#+#+#+#+#+#+# #+#+##+##+#+#+#+#+#+#+#                  
             ##+#+#+##++#+#+ #+##+#+#+#+#+#+#+#+#+                    
                 ##     #+## +#+++#+#+##+#+#+++#                      
                             #++#+##+#+#+#++#+                        
                              +#+#+#+#+#+++#++                        
                              ##+#+#+##+#+#+                          
                               #+##+#+#+##+# .                        
                              +#+# ++## +#+#+                         
                          # ##+ +++##+#+## ##++#                      
                       +#+#+####++#+#+#+#++#+#+#+#+                   
                       #+#+#+ ++#+#+#+#+#+++ #+###+                   
                       ##+#+#-              #+#++#+                   
                                                                      
                                                                      
                                                                      
                                                                      
                                                                      



Curlsmith is a terminal UI HTTP client written in Go.

The idea is a medieval blacksmith forging requests at the anvil: shape a request, temper it, save it into a collection, load it back out, and send it.

## Status

This project is still a work in progress.

What exists today:

- A Bubble Tea TUI with a forge-themed request editor
- Build Request and Load Collection views
- Send HTTP requests
- Save requests into a JSON collection file
- Load stored requests from a collection file
- Vim-like normal/insert navigation inside the TUI

What is not finished yet:

- Collection variables are not fully integrated into request execution
- The collection editing workflow is still basic
- Response inspection is still a single-panel view
- The overall UX and information architecture will continue to change

## Screenshot Mental Model

The TUI is split into two main sections:

- Forge Bench: where you build or load a request
- Anvil Ledger: where status, errors, and responses are shown

The Forge Bench has two views:

- Build Request: method, name, URL, collection path, headers, body
- Load Collection: collection path, summary, stored request list

## Requirements

- Go 1.26.1
- A terminal that works well with Bubble Tea alternate screen mode

## Run

From the repository root:

```bash
go run ./cmd
```

## Test

```bash
go test ./...
```

## Current Controls

### Global

- `ctrl+c`: quit
- `left` / `right`: switch between top-level views when the mode tabs are focused

### Vim-style Modes

- `i`, `a`, `o`: enter INSERT mode on editable fields
- `esc`: return to NORMAL mode
- `q`: quit when in NORMAL mode
- `j`, `k`: move focus in NORMAL mode
- `h`, `l`: move left/right in NORMAL mode where supported

The header shows whether you are in `NORMAL` or `INSERT` mode.

### Build Request View

- `ctrl+s`: send the current request
- `ctrl+w`: save the current request into the selected collection file

### Load Collection View

- `ctrl+r`: load the collection file
- `up` / `down`: move through stored requests
- `enter`: load the selected request back into Build Request

## Collections

Collections are stored as JSON.

Current shape:

```json
{
  "variables": {
    "base_url": "https://api.example.com"
  },
  "requests": [
    {
      "id": "",
      "name": "list items",
      "method": "GET",
      "url": "https://api.example.com/items",
      "headers": {
        "Accept": ["application/json"]
      },
      "body": ""
    }
  ]
}
```

Notes:

- If the target collection file does not exist, saving will create it.
- Saving currently updates an existing stored request by matching `id` first, then `name`.
- If no match is found, the request is appended as a new entry.

## Project Layout

The codebase follows a simple clean-architecture direction:

- [cmd/main.go](/home/mike/Code/curlsmith/cmd/main.go): composition root
- [internal/domain/models.go](/home/mike/Code/curlsmith/internal/domain/models.go): core entities
- [internal/app/send_request.go](/home/mike/Code/curlsmith/internal/app/send_request.go): request sending use case
- [internal/app/load_collection.go](/home/mike/Code/curlsmith/internal/app/load_collection.go): collection loading use case
- [internal/app/save_request_to_collection.go](/home/mike/Code/curlsmith/internal/app/save_request_to_collection.go): collection save use case
- [internal/infra/http/client.go](/home/mike/Code/curlsmith/internal/infra/http/client.go): HTTP adapter
- [internal/infra/storage/file_store.go](/home/mike/Code/curlsmith/internal/infra/storage/file_store.go): JSON file storage adapter
- [internal/tui/model.go](/home/mike/Code/curlsmith/internal/tui/model.go): Bubble Tea terminal interface

The intended dependency direction is:

- `tui` depends on app-layer interfaces/use cases
- `app` depends on domain models and ports
- `infra` implements the ports
- `main` wires everything together

## Design Direction

Curlsmith is not trying to be a generic gray terminal dashboard.

The theme is deliberate:

- forge bench for request creation
- anvil ledger for outputs
- request crafting as a blacksmithing metaphor

That theme will keep shaping the UI as the project evolves.

## Near-Term Roadmap

- Better request and collection management inside the TUI
- Variable resolution during send and preview
- Richer response views for body, headers, and timing
- More robust collection editing semantics
- Additional tests around the TUI workflows