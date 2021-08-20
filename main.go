package main

import (
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"math"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/search"
)

var explainScores = flag.Bool("explain", false, "explain document match scores")

func main() {

	flag.Parse()

	cfg := bluge.InMemoryOnlyConfig()
	w, err := bluge.OpenWriter(cfg)
	if err != nil {
		log.Fatalf("error opening writer: %v", err)
	}
	defer func() {
		_ = w.Close()
	}()

	// index two documents,
	// we are using a stored only field named "boost" to encode a document specific boost
	// the format is simple, and only needs to be known by us when we decode it for
	// custom scoring at search time
	docs := []*bluge.Document{
		// document 'a' has term 'cat' in field 'desc' with freq 3 and a boost of 1.0
		bluge.NewDocument("a").
			AddField(bluge.NewTextField("desc", "cat cat cat")).
			AddField(bluge.NewStoredOnlyField("boost", float64ToBytes(1.0)).
				StoreValue()),

		// document 'b' has term 'cat' in field 'desc' with freq 1 and a boost of 2.0
		bluge.NewDocument("b").
			AddField(bluge.NewTextField("desc", "cat")).
			AddField(bluge.NewStoredOnlyField("boost", float64ToBytes(2.0)).
				StoreValue()),
	}
	for _, doc := range docs {
		err = w.Update(doc.ID(), doc)
		if err != nil {
			log.Fatalf("error updating document '%s': %v", doc.ID(), err)
		}
	}

	// get a reader
	r, err := w.Reader()
	if err != nil {
		log.Fatalf("error getting reader: %v", err)
	}
	defer func() {
		_ = r.Close()
	}()

	// search for cat
	q := bluge.NewTermQuery("cat").SetField("desc")
	req := bluge.NewTopNSearch(10, q)
	if *explainScores {
		req.ExplainScores()
	}
	dmi, err := r.Search(context.TODO(), req)
	if err != nil {
		log.Fatalf("error searching: %v", err)
	}

	// iterate through results
	fmt.Println("natural term score :")
	err = printResults(dmi)
	if err != nil {
		log.Fatalf("error printing results: %v", err)
	}

	// the custom score query will find all matches returned by the provided query q
	// but will also execute a provided transformation function on each DocumentMatch
	// also, by using an anonymous function here, we can execute code on the exact same
	// reader that the search is executing on, and since readers are isolated snapshots
	// this is completely safe.
	// in this case, the transformation we apply is to attempt to load a document boost
	// out of the matched document's stored fields, using a default boost of 1.0 if it
	// is not found.
	// the document match's original score is multiplied by the document boost
	// the document match's explanation is updated to explain this change
	csq := NewCustomScoreQuery(q, func(match *search.DocumentMatch) *search.DocumentMatch {
		if match != nil {
			documentBoost := 1.0
			_ = r.VisitStoredFields(match.Number, func(field string, value []byte) bool {
				if field == "boost" {
					documentBoost = float64FromBytes(value)
				}
				return true
			})

			origScore := match.Score
			match.Score = origScore * documentBoost
			if match.Explanation != nil {
				match.Explanation = search.NewExplanation(match.Score, "custom, doc_boost * orig_score",
					search.NewExplanation(documentBoost, "doc_boost, loaded from field 'boost'"),
					match.Explanation)
			}
		}
		return match
	})

	csreq := bluge.NewTopNSearch(10, csq)
	if *explainScores {
		csreq.ExplainScores()
	}
	dmi, err = r.Search(context.TODO(), csreq)
	if err != nil {
		log.Fatalf("error searching: %v", err)
	}

	// iterate through results
	fmt.Println("custom score with document boost:")
	err = printResults(dmi)
	if err != nil {
		log.Fatalf("error printing results: %v", err)
	}
}

func printResults(dmi search.DocumentMatchIterator) error {
	next, err := dmi.Next()
	for next != nil && err == nil {
		sfErr := next.VisitStoredFields(func(field string, value []byte) bool {
			if field == "_id" {
				fmt.Printf("id: %s score: %f expl: %s\n", string(value), next.Score, next.Explanation)
			}
			return true
		})
		if sfErr != nil {
			log.Fatalf("error loading stored fields: %v", err)
		}

		next, err = dmi.Next()
	}
	return err
}

func float64ToBytes(f float64) (rv []byte) {
	rv = make([]byte, 8)
	bits := math.Float64bits(f)
	binary.BigEndian.PutUint64(rv, bits)
	return rv
}

func float64FromBytes(b []byte) float64 {
	bits := binary.BigEndian.Uint64(b)
	return math.Float64frombits(bits)
}
