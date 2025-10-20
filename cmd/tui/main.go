package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"

	"github.com/dukerupert/dd/data/sql/migrations"
	"github.com/dukerupert/dd/internal/store"
)

type viewState int

const (
	viewList viewState = iota
	viewPromptAlbum
	viewPromptArtist
	viewPromptYear
	viewPromptLocation
)

type model struct {
	queries   *store.Queries
	records   []store.Record
	artists   []store.Artist
	locations []store.Location
	cursor    int
	err       error
	loaded    bool
	state     viewState

	// Input state
	input string

	// Form data being collected
	newRecordTitle      string
	newRecordArtist     string
	newRecordArtistID   *int64
	newRecordAlbum      string
	newRecordYear       string
	newRecordLocation   string
	newRecordLocationID *int64

	// Filtered results for artist/location search
	filteredArtists   []store.Artist
	filteredLocations []store.Location
	selectionCursor   int
}

type recordsLoadedMsg struct {
	records []store.Record
	err     error
}

func loadRecordsCmd(queries *store.Queries) tea.Cmd {
	return func() tea.Msg {
		records, err := queries.ListRecords(context.Background())
		return recordsLoadedMsg{records: records, err: err}
	}
}

type artistsLoadedMsg struct {
	artists []store.Artist
	err     error
}

func loadArtistsCmd(queries *store.Queries) tea.Cmd {
	return func() tea.Msg {
		artists, err := queries.ListArtists(context.Background())
		return artistsLoadedMsg{artists: artists, err: err}
	}
}

func (m model) Init() tea.Cmd {
	return loadRecordsCmd(m.queries)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case recordsLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		m.records = msg.records
		m.loaded = true
		return m, nil

	case artistsLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		m.artists = msg.artists
		m.filteredArtists = msg.artists
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case viewList:
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.records)-1 {
					m.cursor++
				}
			case "n":
				m.state = viewPromptAlbum // Changed from viewPromptTitle
				m.input = ""
				return m, nil
			}

		case viewPromptArtist:
			switch msg.String() {
			case "enter":
				// TODO: Handle artist selection/creation
				m.newRecordArtist = m.input
				m.input = ""
				m.state = viewPromptAlbum
				return m, nil
			case "esc":
				m.state = viewList
				return m, nil
			default:
				m.input += msg.String()
				// TODO: Filter m.filteredArtists based on m.input
			}

		case viewPromptAlbum:
			switch msg.String() {
			case "enter":
				m.newRecordAlbum = m.input
				m.input = ""
				m.state = viewPromptArtist
				return m, nil
			case "esc":
				m.state = viewList
				return m, nil
			default:
				m.input += msg.String()
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	if !m.loaded {
		return "Loading records...\n"
	}

	switch m.state {

	case viewList:
		if len(m.records) == 0 {
			return "No records found.\n\nPress 'n' to create a new record\nPress 'q' to quit\n"
		}

		s := "Vinyl Collection\n\n"

		for i, record := range m.records {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			s += fmt.Sprintf("%s %s - %s\n", cursor, record.Title, record.AlbumTitle.String)
		}

		s += "\nPress 'n' for new record, 'q' to quit.\n"
		return s

	case viewPromptArtist:
		return fmt.Sprintf("Enter artist name: %s", m.input)

	case viewPromptAlbum:
		return fmt.Sprintf("Enter album title: %s", m.input)
	}

	return ""
}

func initDB(dbPath string) (*store.Queries, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	provider, err := goose.NewProvider(goose.DialectSQLite3, db, migrations.Embed)
	if err != nil {
		return nil, err
	}

	if _, err := provider.Up(context.Background()); err != nil {
		return nil, err
	}

	return store.New(db), nil
}

func main() {
	queries, err := initDB("./data/vinyl.db")
	if err != nil {
		fmt.Println("Database error:", err)
		os.Exit(1)
	}

	m := model{queries: queries}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
