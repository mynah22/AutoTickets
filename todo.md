# to do

- if api calls fail, push something to ui
  - use heartbeat to determine last successful api call and if server is sleeping (outside active hours)
- find and appropriately handle ALL error returns
- compare index.html
  - standardize nomenclature and syntax/style
  - write helper functions (particularly for visibility changes)
- go func when ranging wsClients for broadcast
  - parallel update messages
- implement resource count logic
  - implement UI to edit and save ResourceID:name map
  - return rescId counts in existing tickets message broadcast
