package transformer

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

const startHighlightTag string = "<strong>"
const endHighlightTag string = "</strong>"

// Transformer represents an instance of the ResponseTransformer interface
type Transformer struct{}

// Structs representing the transformed response
type searchResponse struct {
	Count               int           `json:"count"`
	Took                int           `json:"took"`
	ContentTypes        []contentType `json:"content_types"`
	Items               []contentItem `json:"items"`
	Suggestions         []string      `json:"suggestions,omitempty"`
	AdditionSuggestions []string      `json:"additional_suggestions,omitempty"`
}

type contentType struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

type contentItem struct {
	Description description `json:"description"`
	Type        string      `json:"type"`
	URI         string      `json:"uri"`
	Matches     *matches    `json:"matches,omitempty"`
}

type description struct {
	Contact           *contact  `json:"contact,omitempty"`
	DatasetID         string    `json:"dataset_id,omitempty"`
	Edition           string    `json:"edition,omitempty"`
	Headline1         string    `json:"headline1,omitempty"`
	Headline2         string    `json:"headline2,omitempty"`
	Headline3         string    `json:"headline3,omitempty"`
	Keywords          *[]string `json:"keywords,omitempty"`
	LatestRelease     *bool     `json:"latest_release,omitempty"`
	Language          string    `json:"language,omitempty"`
	MetaDescription   string    `json:"meta_description,omitempty"`
	NationalStatistic *bool     `json:"national_statistic,omitempty"`
	NextRelease       string    `json:"next_release,omitempty"`
	PreUnit           string    `json:"pre_unit,omitempty"`
	ReleaseDate       string    `json:"release_date,omitempty"`
	Source            string    `json:"source,omitempty"`
	Summary           string    `json:"summary"`
	Title             string    `json:"title"`
	Unit              string    `json:"unit,omitempty"`
}

type contact struct {
	Name      string `json:"name"`
	Telephone string `json:"telephone,omitempty"`
	Email     string `json:"email"`
}

type matches struct {
	Description struct {
		Summary         *[]matchDetails `json:"summary"`
		Title           *[]matchDetails `json:"title"`
		Edition         *[]matchDetails `json:"edition,omitempty"`
		MetaDescription *[]matchDetails `json:"meta_description,omitempty"`
		Keywords        *[]matchDetails `json:"keywords,omitempty"`
		DatasetID       *[]matchDetails `json:"dataset_id,omitempty"`
	} `json:"description"`
}

type matchDetails struct {
	Value string `json:"value,omitempty"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

// Structs representing the raw elastic search response

type esResponse struct {
	Responses []esResponseItem `json:"responses"`
}

type esResponseItem struct {
	Took         int                    `json:"took"`
	Hits         esResponseHits         `json:"hits"`
	Aggregations esResponseAggregations `json:"aggregations"`
	Suggest      esSuggest              `json:"suggest"`
}

type esResponseHits struct {
	Total int
	Hits  []esResponseHit `json:"hits"`
}

type esResponseHit struct {
	Source    esSourceDocument `json:"_source"`
	Highlight esHighlight      `json:"highlight"`
}

type esSourceDocument struct {
	Description struct {
		Summary           string    `json:"summary"`
		NextRelease       string    `json:"nextRelease,omitempty"`
		Unit              string    `json:"unit,omitempty"`
		Keywords          *[]string `json:"keywords,omitempty"`
		ReleaseDate       string    `json:"releaseDate,omitempty"`
		Edition           string    `json:"edition,omitempty"`
		LatestRelease     *bool     `json:"latestRelease,omitempty"`
		Language          string    `json:"language,omitempty"`
		Contact           *contact  `json:"contact,omitempty"`
		DatasetID         string    `json:"datasetId,omitempty"`
		Source            string    `json:"source,omitempty"`
		Title             string    `json:"title"`
		MetaDescription   string    `json:"metaDescription,omitempty"`
		NationalStatistic *bool     `json:"nationalStatistic,omitempty"`
		PreUnit           string    `json:"preUnit,omitempty"`
		Headline1         string    `json:"headline1,omitempty"`
		Headline2         string    `json:"headline2,omitempty"`
		Headline3         string    `json:"headline3,omitempty"`
	} `json:"description"`
	Type string `json:"type"`
	URI  string `json:"uri"`
}

type esHighlight struct {
	DescriptionTitle     *[]string `json:"description.title"`
	DescriptionEdition   *[]string `json:"description.edition"`
	DescriptionSummary   *[]string `json:"description.summary"`
	DescriptionMeta      *[]string `json:"description.metaDescription"`
	DescriptionKeywords  *[]string `json:"description.keywords"`
	DescriptionDatasetID *[]string `json:"description.datasetId"`
}

type esResponseAggregations struct {
	DocCounts struct {
		Buckets []esBucket `json:"buckets"`
	} `json:"docCounts"`
}

type esBucket struct {
	Key   string `json:"key"`
	Count int    `json:"doc_count"`
}

type esSuggest struct {
	SearchSuggest []esSearchSuggest `json:"search_suggest"`
}

type esSearchSuggest struct {
	Options []esSearchSuggestOptions `json:"options"`
}

type esSearchSuggestOptions struct {
	Text string `json:"text"`
}

// New returns a new instance of Transformer
func New() *Transformer {
	return &Transformer{}
}

// TransformSearchResponse transforms an elastic search response into a structure that matches the v1 api specification
func (t *Transformer) TransformSearchResponse(ctx context.Context, responseData []byte, query string) ([]byte, error) {
	var source esResponse

	err := json.Unmarshal(responseData, &source)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to decode elastic search response")
	}

	if len(source.Responses) < 1 {
		return nil, errors.New("Response to be transformed contained 0 items")
	}

	sr := transform(&source)

	if sr.Count == 0 {
		as := buildAdditionalSuggestionList(query)
		sr.AdditionSuggestions = as
	}

	transformedData, err := json.Marshal(sr)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to encode transformed response")
	}
	return transformedData, nil
}

func transform(source *esResponse) searchResponse {
	sr := searchResponse{
		Count:        source.Responses[0].Hits.Total,
		Items:        []contentItem{},
		ContentTypes: []contentType{},
	}
	var took int = 0
	for _, response := range source.Responses {
		for _, doc := range response.Hits.Hits {
			sr.Items = append(sr.Items, buildContentItem(doc))
		}
		for _, bucket := range response.Aggregations.DocCounts.Buckets {
			sr.ContentTypes = append(sr.ContentTypes, buildContentTypes(bucket))
		}
		for _, suggest := range response.Suggest.SearchSuggest {
			for _, option := range suggest.Options {
				sr.Suggestions = append(sr.Suggestions, option.Text)
			}
		}
		took += response.Took
	}
	sr.Took = took
	return sr
}

func buildContentItem(doc esResponseHit) contentItem {
	ci := contentItem{
		Description: buildDescription(doc),
		Type:        doc.Source.Type,
		URI:         doc.Source.URI,
		Matches:     buildMatches(doc.Highlight),
	}

	return ci
}

func buildDescription(doc esResponseHit) description {
	sd := doc.Source.Description
	return description{
		Summary:           sd.Summary,
		NextRelease:       sd.NextRelease,
		Unit:              sd.Unit,
		PreUnit:           sd.PreUnit,
		Keywords:          sd.Keywords,
		ReleaseDate:       sd.ReleaseDate,
		Edition:           sd.Edition,
		LatestRelease:     sd.LatestRelease,
		Language:          sd.Language,
		Contact:           sd.Contact,
		DatasetID:         sd.DatasetID,
		Source:            sd.Source,
		Title:             sd.Title,
		MetaDescription:   sd.MetaDescription,
		NationalStatistic: sd.NationalStatistic,
		Headline1:         sd.Headline1,
		Headline2:         sd.Headline2,
		Headline3:         sd.Headline3,
	}
}

func buildMatches(hl esHighlight) *matches {
	var matches matches

	if highlights := hl.DescriptionTitle; highlights != nil {
		var titleMatches []matchDetails
		for _, m := range *highlights {
			foundMatchDetails, _ := findMatches(m)
			titleMatches = append(titleMatches, foundMatchDetails...)
		}
		matches.Description.Title = &titleMatches
	}

	if highlights := hl.DescriptionEdition; highlights != nil {
		var editionMatches []matchDetails
		for _, m := range *highlights {
			foundMatchDetails, _ := findMatches(m)
			editionMatches = append(editionMatches, foundMatchDetails...)
		}
		matches.Description.Edition = &editionMatches
	}

	if highlights := hl.DescriptionSummary; highlights != nil {
		var summaryMatches []matchDetails
		for _, m := range *highlights {
			foundMatchDetails, _ := findMatches(m)
			summaryMatches = append(summaryMatches, foundMatchDetails...)
		}
		matches.Description.Summary = &summaryMatches
	}

	if highlights := hl.DescriptionMeta; highlights != nil {
		var summaryMatches []matchDetails
		for _, m := range *highlights {
			foundMatchDetails, _ := findMatches(m)
			summaryMatches = append(summaryMatches, foundMatchDetails...)
		}
		matches.Description.MetaDescription = &summaryMatches
	}

	if highlights := hl.DescriptionKeywords; highlights != nil {
		var keywordsMatches []matchDetails
		for _, m := range *highlights {
			foundMatchDetails, value := findMatches(m)
			for _, md := range foundMatchDetails {
				md.Value = value
				keywordsMatches = append(keywordsMatches, md)
			}
		}
		matches.Description.Keywords = &keywordsMatches
	}

	if highlights := hl.DescriptionDatasetID; highlights != nil {
		var datasetIDMatches []matchDetails
		for _, m := range *highlights {
			foundMatchDetails, _ := findMatches(m)
			datasetIDMatches = append(datasetIDMatches, foundMatchDetails...)
		}
		matches.Description.DatasetID = &datasetIDMatches
	}

	return &matches
}

// Find matches finds all the matching marked-up phrases and returns a slice of their start and end points in the string
// NB. The start and end values are the number of bytes, not characters, so be aware when the input contains higher-order
// UTF-8 characters.
func findMatches(s string) ([]matchDetails, string) {

	md := make([]matchDetails, 0, 2)
	fs := s

	if start := strings.Index(s, startHighlightTag); start >= 0 {
		left := s[0:start]
		right := s[start+len(startHighlightTag) : len(s)]
		if end := strings.Index(right, endHighlightTag); end >= 0 {
			mid := right[0:end]
			remain := right[end+len(endHighlightTag) : len(right)]

			md = append(md, matchDetails{
				Start: start + 1,
				End:   start + end,
			})

			remainMatches, remain := findMatches(remain)
			for _, rm := range remainMatches {
				rm.Start += len(left) + len(mid)
				rm.End += len(left) + len(mid)
				md = append(md, rm)
			}

			right = mid + remain
		}

		fs = left + right
	}

	return md, fs
}

func buildContentTypes(bucket esBucket) contentType {
	return contentType{
		Type:  bucket.Key,
		Count: bucket.Count,
	}
}

func buildAdditionalSuggestionList(query string) []string {
	regex := regexp.MustCompile(`"[^"]*"|\S+`)

	queryTerms := []string{}
	for _, match := range regex.FindAllStringSubmatch(query, -1) {
		queryTerms = append(queryTerms, match[0])
	}
	return queryTerms
}
