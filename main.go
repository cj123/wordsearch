package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/go-chi/chi"
	"github.com/jung-kurt/gofpdf"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func main() {
	r := chi.NewRouter()

	t, err := template.New("wordsearch").Funcs(sprig.FuncMap()).Parse(htmlTemplate)

	if err != nil {
		panic(err)
	}

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		err := t.Execute(w, map[string]interface{}{
			"NumWords": numWords,
		})

		if err != nil {
			panic(err)
		}
	})

	r.Get("/wordsearch.pdf", func(w http.ResponseWriter, r *http.Request) {
		ws := WordSearch{}

		for i := 0; i < numWords; i++ {
			ws.Words[i] = r.URL.Query().Get(fmt.Sprintf("Word_%d", i))
		}

		w.Header().Add("Content-Type", "application/pdf")
		w.Header().Add("Content-Disposition", `inline; filename="wordsearch.pdf"`)

		ws.Generate()
		ws.PDF(w)
	})

	log.Println("Listening on: :5598")
	http.ListenAndServe(":5598", r)
}

const (
	rows    = 12
	columns = 14

	numWords = 14
)

type WordSearch struct {
	Words [numWords]string

	output [rows][columns]string
}

func (ws *WordSearch) Generate() {
	// find our starting positions
	for _, word := range ws.Words {
		foundSpace := false

		var rowIndex, columnIndex int

		for !foundSpace {
			rowIndex = rand.Intn(rows)
			columnIndex = rand.Intn(columns - len(word))

			row := ws.output[rowIndex]

			allSpacesFree := true

			for i := 0; i < len(word); i++ {
				if row[columnIndex+i] != "" {
					allSpacesFree = false
				}
			}

			if allSpacesFree {
				foundSpace = true
			}
		}

		// place the word in the search
		for index, letter := range word {
			ws.output[rowIndex][columnIndex+index] = strings.ToUpper(fmt.Sprintf("%c", letter))
		}
	}

	for rowIndex, row := range ws.output {
		for columnIndex, letter := range row {
			if letter == "" {
				ws.output[rowIndex][columnIndex] = strings.ToUpper(fmt.Sprintf("%c", letters[rand.Intn(len(letters))]))
			}
		}
	}
}

const letters = "abcdefghijklmnopqrstuvwxyj"

func (ws *WordSearch) Print() {
	for _, row := range ws.output {
		for _, letter := range row {
			fmt.Printf("%s ", letter)
		}

		fmt.Println()
	}

	fmt.Println()

	for wordIndex, word := range ws.Words {
		fmt.Printf("%d. %s\n", wordIndex+1, word)
	}
}

func (ws *WordSearch) PDF(w io.Writer) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle("Word Search", true)

	const columnSize = 10

	makeTable := func() {
		left := (210.0 - columns*columnSize) / 2
		pdf.SetX(left)
		for _, row := range ws.output {
			pdf.SetX(left)

			for _, letter := range row {
				pdf.SetFontSize(20)
				pdf.CellFormat(columnSize, columnSize, letter, "1", 0, "C", false, 0, "")
			}

			pdf.Ln(-1)
		}
	}
	pdf.SetFont("Arial", "", 14)
	pdf.AddPage()

	makeTable()

	pdf.Ln(-1)
	pdf.Ln(-1)

	currentCol := 0
	yAfterTable := pdf.GetY()

	pdf.SetAcceptPageBreakFunc(func() bool {
		if currentCol > 0 {
			return true
		}

		currentCol++

		x := 10.0 + float64(currentCol)*100.0

		pdf.SetLeftMargin(x)
		pdf.SetX(x)
		pdf.SetY(yAfterTable)

		return false
	})

	pdf.SetLeftMargin(30)

	for i, word := range ws.Words {
		pdf.Cell(0, 14, fmt.Sprintf("%d. %s", i+1, word))
		pdf.Ln(-1)
	}

	err := pdf.Output(w)
	return err
}

const htmlTemplate = `
<!doctype html>
<html lang="en">
<head>
	<title>Word Search</title>
	<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/twitter-bootstrap/4.1.3/css/bootstrap.min.css" />
</head>
<body>
	<nav class="navbar navbar-expand-lg navbar-dark bg-dark">
		<div class="container">
			<a class="navbar-brand" href="/">Word Search</a>
		</div>
	</nav>
	<div class="container">
		<h1 class="text-center mt-3 mb-3">Word Search</h1>

		<p>Add your words below, and then click "Make Word Search!". You should then be able to print the wordsearch from the next screen! You can leave some words blank if you want.</p>

		<form method="GET" action="wordsearch.pdf">
			{{ range $index, $count := until $.NumWords }}
				<div class="form-group row">
					<label for="Word_{{ $count }}" class="col-sm-2 col-form-label">Word {{ add $count 1 }}</label>
					<div class="col-sm-10">
						<input type="text" class="form-control" name="Word_{{ $count }}">
					</div>
				</div>
			{{ end }}

			<button type="submit" class="btn btn-success float-right">Make Word Search!</button>
		</form>

		<div class="clearfix"></div>
	</div>
	<footer class="text-muted mt-5 pt-5 pb-5 mb-5">
		<div class="container">
			<em>&copy; 2020 Callum Jones</em>
		</div>
	</footer>
</body>
</html>
`
