package main

import (
	"github.com/b5160r/curlsmith/internal/app"
	httpinfra "github.com/b5160r/curlsmith/internal/infra/http"
	storageinfra "github.com/b5160r/curlsmith/internal/infra/storage"
	"github.com/b5160r/curlsmith/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	collectionStore := storageinfra.CollectionStore{}
	useCase := &app.SendRequestUseCase{Client: &httpinfra.Client{}}
	loadCollection := &app.LoadCollectionUseCase{Loader: collectionStore}
	saveRequest := &app.SaveRequestToCollectionUseCase{Store: collectionStore}

	p := tea.NewProgram(tui.NewModel(useCase, loadCollection, saveRequest), tea.WithAltScreen())
	if err := p.Start(); err != nil {
		panic(err)
	}
}
