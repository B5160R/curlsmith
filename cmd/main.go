package main

import (
	"github.com/b5160r/curlsmith/internal/app"
	httpinfra "github.com/b5160r/curlsmith/internal/infra/http"
	storageinfra "github.com/b5160r/curlsmith/internal/infra/storage"
	"github.com/b5160r/curlsmith/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	store := storageinfra.FileStore{}
	useCase := &app.SendRequestUseCase{Client: &httpinfra.Client{}}
	loadCollection := &app.LoadCollectionUseCase{Loader: store}
	saveRequest := &app.SaveRequestToCollectionUseCase{Store: store}

	p := tea.NewProgram(tui.NewModel(useCase, loadCollection, saveRequest), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
