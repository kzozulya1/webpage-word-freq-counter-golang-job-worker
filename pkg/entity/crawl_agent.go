package entity

import (
	pb "github.com/kzozulya1/webpage-word-freq-counter-protobuf/protobuf"
	"github.com/denisbrodbeck/striphtmltags"
	"io/ioutil"
	"net/http"
	"regexp"
	"sort"
	"time"
	//logger "app/pkg/loggerutil"
	//"github.com/PuerkitoBio/goquery"
)

//Entity with no data, but with behavior
type CrawlAgent struct {
	
}

//For sorting purposes
type KeyValueType struct {
	Key   string
	Value int
}

//Parse url
//get content of current page, count words
//recursively process next links
func (c *CrawlAgent) Process(job *Job) (*pb.PageWordFrequency, error) {
	pageContent, err := c.getPageContent(job.GetURL()) 
	if err != nil {
		return nil, err
	}

	pageTitle := c.parsePageTitle(string(pageContent))

	var strippedContents string = c.stripHtmlTags(pageContent)
	strippedContents, err = c.skipSpecialChars(strippedContents)
	if err != nil {
		return nil, err
	}
	var countedWords map[string]int = c.countWords(strippedContents)
	c.applyFilter(&countedWords, job)

	filter := c.prepareAppliedFilter(job)

	//fmt.Printf("New data: %v",countedWords)
	// logger.Log(countedWords,"result.log")
	// logger.Log("\n\n~~~\n\n","result.log")
	return c.preparePageWordFrequency(pageTitle,job.GetURL(), filter, &countedWords), nil
}

//Get website page content
func (c *CrawlAgent) getPageContent(url string) ([]byte, error) {
	resp, err := http.Get(url)
	// handle the error if there is one
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

//Strip tags
func (c *CrawlAgent) stripHtmlTags(data []byte) string {
	return striphtmltags.StripTags(string(data))
}

//Strip special chars
func (c *CrawlAgent) skipSpecialChars(s string) (string, error) {
	//remove all not symbols
	reg, err := regexp.Compile("[^\\s\\p{L}]+")
	if err != nil {
		return s, err
	}
	return reg.ReplaceAllString(s, ""), nil 
}

// WordCount returns a map of the counts of each “word” in the string s.
func (c *CrawlAgent) countWords(s string) map[string]int {
	//words := strings.Fields(s)
	
	var words []string = regexp.MustCompile(`\s+`).Split(s,-1)

	counts := make(map[string]int, len(words))
	for _, word := range words {
		counts[word]++
	}
	return counts
}

//Apply extra filter on results
func (c *CrawlAgent) applyFilter(data *map[string]int, job *Job) {
	//Check exclude list filter
	for k, v := range *data {
		//Exclude list filter check
		if job.GetFilter().GetExcludeList() != nil {
			for _, exclWord := range job.GetFilter().GetExcludeList() {
				if k == exclWord {
					delete(*data, k)
				}
			}
		}
		//Min freq filter check
		if job.GetFilter().GetMinFrequency() > 0 && v < job.GetFilter().GetMinFrequency() {
			delete(*data, k)
		}
		//Min len word filter check
		if job.GetFilter().GetMinLen() > 0 && c.getTermLen(k) < job.GetFilter().GetMinLen() {
			delete(*data, k)
		}
	}
}

//Get len of term: non en occupies 2 bytes. En only 1 byte
func (c *CrawlAgent) getTermLen(term string) int {
	isEng := regexp.MustCompile(`[a-zA-Z]`)
	tLen := len(term)
	if !isEng.Match([]byte(term)){
		tLen /= 2
	}
	return tLen
}

//Extract title
func (c *CrawlAgent) parsePageTitle(data string) string {
	result := ""
	re := regexp.MustCompile(`<title.*>(.*)</title>`) 
        if re.Match([]byte(data)) {
			result = re.FindStringSubmatch(data)[1]
	}
	return result
}

//Prepare applied filter
func (c *CrawlAgent) prepareAppliedFilter(job *Job) *pb.AppliedFilter {
	return &pb.AppliedFilter{
			MinFrequency: int32(job.GetFilter().GetMinFrequency()),
			MinLen: int32(job.GetFilter().GetMinLen()),
			ExcludeList: job.GetFilter().GetExcludeList(),
		}
}

//Prepare gRPC message
func (c *CrawlAgent) preparePageWordFrequency(title string, url string, af *pb.AppliedFilter ,countedWords *map[string]int) *pb.PageWordFrequency {
	res := pb.PageWordFrequency{PageUrl: url, PageTitle: title, Words: make([]*pb.Word, 1, 100), UpdatedAt: int32(time.Now().Unix()), AppliedFilter: af}
	
	//Prepare Words subarray, sort on freq
	var ss []KeyValueType
    for k, v := range *countedWords {
        ss = append(ss, KeyValueType{k, v})
    }
    sort.Slice(ss, func(i, j int) bool {
        return ss[i].Value > ss[j].Value
    })
    for _, kv := range ss {
		res.Words = append(res.Words, &pb.Word{Value: kv.Key, Count: int32(kv.Value)})
    }
	return &res
}