package microdata

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseItemScope(t *testing.T) {
	html := `
		<div itemscope itemtype="https://example.com/Person">
			<p>My name is <span itemprop="name">Penelope</span>.</p>
		</div>`

	data := ParseData(html, t)

	result := len(data.Items[0].Properties)
	expected := 1
	if result != expected {
		t.Errorf("Result should have been \"%d\", but it was \"%d\"", expected, result)
	}
}

func TestParseItemType(t *testing.T) {
	html := `
		<div itemscope itemtype="https://example.com/Person">
			<p>My name is <span itemprop="name">Penelope</span>.</p>
		</div>`

	data := ParseData(html, t)

	result := data.Items[0].Types[0]
	expected := "https://example.com/Person"
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func TestParseItemRef(t *testing.T) {
	html := `
		<div itemscope itemtype="https://example.com/Movie" itemref="properties">
			<p><span itemprop="name">Rear Window</span> is a movie from 1954.</p>
		</div>
		<ul id="properties">
			<li itemprop="genre">Thriller</li>
			<li itemprop="description">A homebound photographer spies on his neighbours.</li>
		</ul>`

	var testTable = []struct {
		propName string
		expected string
	}{
		{"genre", "Thriller"},
		{"description", "A homebound photographer spies on his neighbours."},
	}

	data := ParseData(html, t)

	for _, test := range testTable {
		if result := data.Items[0].Properties[test.propName][0].(string); result != test.expected {
			t.Errorf("Result should have been \"%s\", but it was \"%s\"", test.expected, result)
		}
	}
}

func TestParseItemProp(t *testing.T) {
	html := `
		<div itemscope itemtype="https://example.com/Person">
			<p>My name is <span itemprop="name">Penelope</span>.</p>
		</div>`

	data := ParseData(html, t)

	result := data.Items[0].Properties["name"][0].(string)
	expected := "Penelope"
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func TestParseItemId(t *testing.T) {
	html := `
		<ul itemscope itemtype="https://example.com/Book" itemid="urn:isbn:978-0141196404">
			<li itemprop="title">The Black Cloud</li>
			<li itemprop="author">Fred Hoyle</li>
		</ul>`

	data := ParseData(html, t)

	result := data.Items[0].ID
	expected := "urn:isbn:978-0141196404"
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func TestParseHref(t *testing.T) {
	html := `
		<html itemscope itemtype="https://example.com/Person">
			<head>
				<link itemprop="linkTest" href="https://example.com/cde">
			<head>
			<div>
				<a itemprop="aTest" href="https://example.com/abc" /></a>
				<area itemprop="areaTest" href="https://example.com/bcd" />
			</div>
		</div>`

	var testTable = []struct {
		propName string
		expected string
	}{
		{"aTest", "https://example.com/abc"},
		{"areaTest", "https://example.com/bcd"},
		{"linkTest", "https://example.com/cde"},
	}

	data := ParseData(html, t)

	for _, test := range testTable {
		if result := data.Items[0].Properties[test.propName][0].(string); result != test.expected {
			t.Errorf("Result should have been \"%s\", but it was \"%s\"", test.expected, result)
		}
	}
}

func TestParseSrc(t *testing.T) {
	html := `
		<div itemscope itemtype="https://example.com/Videocast">
			<audio itemprop="audioTest" src="https://example.com/abc" />
			<embed itemprop="embedTest" src="https://example.com/bcd" />
			<iframe itemprop="iframeTest" src="https://example.com/cde"></iframe>
			<img itemprop="imgTest" src="https://example.com/def" />
			<source itemprop="sourceTest" src="https://example.com/efg" />
			<track itemprop="trackTest" src="https://example.com/fgh" />
			<video itemprop="videoTest" src="https://example.com/ghi" />
		</div>`

	var testTable = []struct {
		propName string
		expected string
	}{
		{"audioTest", "https://example.com/abc"},
		{"embedTest", "https://example.com/bcd"},
		{"iframeTest", "https://example.com/cde"},
		{"imgTest", "https://example.com/def"},
		{"sourceTest", "https://example.com/efg"},
		{"trackTest", "https://example.com/fgh"},
		{"videoTest", "https://example.com/ghi"},
	}

	data := ParseData(html, t)

	for _, test := range testTable {
		if result := data.Items[0].Properties[test.propName][0].(string); result != test.expected {
			t.Errorf("Result should have been \"%s\", but it was \"%s\"", test.expected, result)
		}
	}
}

func TestParseMetaContent(t *testing.T) {
	html := `
		<html itemscope itemtype="https://example.com/Person">
			<meta itemprop="length" content="1.70" />
		</html>`

	data := ParseData(html, t)

	result := data.Items[0].Properties["length"][0].(string)
	expected := "1.70"
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func TestParseValue(t *testing.T) {
	html := `
		<div itemscope itemtype="https://example.com/Container">
			<data itemprop="capacity" value="80">80 liters</data>
			<meter itemprop="volume" min="0" max="100" value="25">25%</meter>
		</div>`

	var testTable = []struct {
		propName string
		expected string
	}{
		{"capacity", "80"},
		{"volume", "25"},
	}

	data := ParseData(html, t)

	for _, test := range testTable {
		if result := data.Items[0].Properties[test.propName][0].(string); result != test.expected {
			t.Errorf("Result should have been \"%s\", but it was \"%s\"", test.expected, result)
		}
	}
}

func TestParseDatetime(t *testing.T) {
	html := `
		<div itemscope itemtype="https://example.com/Person">
			<time itemprop="birthDate" datetime="1993-10-02">22 years</time>
		</div>`

	data := ParseData(html, t)

	result := data.Items[0].Properties["birthDate"][0].(string)
	expected := "1993-10-02"
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func TestParseText(t *testing.T) {
	html := `
		<div itemscope itemtype="https://example.com/Product">
			<span itemprop="price">3.95</span>
		</div>`

	data := ParseData(html, t)

	result := data.Items[0].Properties["price"][0].(string)
	expected := "3.95"
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func TestParseMultiItemTypes(t *testing.T) {
	html := `
		<div itemscope itemtype="https://example.com/Park https://example.com/Zoo">
			<span itemprop="name">ZooParc Overloon</span>
		</div>`

	data := ParseData(html, t)

	result := len(data.Items[0].Types)
	expected := 2
	if result != expected {
		t.Errorf("Result should have been \"%d\", but it was \"%d\"", expected, result)
	}
}

func TestJSON(t *testing.T) {
	html := `
		<div itemscope itemtype="https://example.com/Person">
			<p>My name is <span itemprop="name">Penelope</span>.</p>
			<p>I am <date itemprop="age" value="22">22 years old.</span>.</p>
		</div>`

	data := ParseData(html, t)

	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	result := string(b)
	expected := `{"items":[{"type":["https://example.com/Person"],"properties":{"age":["22 years old.."],"name":["Penelope"]}}]}`
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func TestParseHTML(t *testing.T) {
	buf := bytes.NewBufferString(gallerySnippet)
	_, result := ParseHTML(buf, "charset=utf-8", "https://blog.example.com/progress-report")
	if result != nil {
		t.Errorf("Result should have been nil, but it was \"%s\"", result)
	}
}

func TestParseURL(t *testing.T) {
	html := `
		<div itemscope itemtype="https://example.com/Person">
			<p>My name is <span itemprop="name">Penelope</span>.</p>
		</div>`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(html))
	}))
	defer ts.Close()

	data, err := ParseURL(ts.URL)
	if err != nil {
		t.Error(err)
	}

	result := data.Items[0].Properties["name"][0].(string)
	expected := "Penelope"
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func TestNestedItems(t *testing.T) {
	html := `
		<div>
			<div itemscope itemtype="https://example.com/Person">
				<p>My name is <span itemprop="name">Penelope</span>.</p>
				<p>I am <date itemprop="age" value="22">22 years old.</span>.</p>
				<div itemscope itemtype="https://example.com/Breadcrumb">
					<a itemprop="url" href="https://example.com/users/1"><span itemprop="title">profile</span></a>
				</div>
			</div>
		</div>`

	data := ParseData(html, t)

	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	result := string(b)
	expected := `{"items":[{"type":["https://example.com/Person"],"properties":{"age":["22 years old.."],"name":["Penelope"]}},{"type":["https://example.com/Breadcrumb"],"properties":{"title":["profile"],"url":["https://example.com/users/1"]}}]}`
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func ParseData(html string, t *testing.T) *Microdata {
	r := strings.NewReader(html)

	data, err := ParseHTML(r, "charset=utf-8", "https://example.com")
	if err != nil {
		t.Error(err)
	}
	return data
}

func TestParseW3CBookSnippet(t *testing.T) {
	buf := bytes.NewBufferString(bookSnippet)
	data, err := ParseHTML(buf, "charset=utf-8", "")
	if err != nil {
		t.Error(err)
	}

	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	result := string(b)
	expected := `{"items":[{"type":["https://vocab.example.net/book"],"properties":{"author":["Peter F. Hamilton"],"pubdate":["1996-01-26"],"title":["The Reality Dysfunction"]},"id":"urn:isbn:0-330-34032-8"}]}`
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func TestParseW3CGalleySnippet(t *testing.T) {
	buf := bytes.NewBufferString(gallerySnippet)
	data, err := ParseHTML(buf, "charset=utf-8", "")
	if err != nil {
		t.Error(err)
	}

	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	result := string(b)
	expected := `{"items":[{"type":["https://n.whatwg.org/work"],"properties":{"license":["https://www.opensource.org/licenses/mit-license.php"],"title":["The house I found."],"work":["/images/house.jpeg"]}},{"type":["https://n.whatwg.org/work"],"properties":{"license":["https://www.opensource.org/licenses/mit-license.php"],"title":["The mailbox."],"work":["/images/mailbox.jpeg"]}}]}`
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func TestParseW3CBlogSnippet(t *testing.T) {
	buf := bytes.NewBufferString(blogSnippet)
	data, err := ParseHTML(buf, "charset=utf-8", "https://blog.example.com/progress-report")
	if err != nil {
		t.Error(err)
	}

	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	result := string(b)
	expected := `{"items":[{"type":["https://schema.org/BlogPosting"],"properties":{"comment":[{"type":["https://schema.org/UserComments"],"properties":{"commentTime":["2013-08-29"],"creator":[{"type":["https://schema.org/Person"],"properties":{"name":["Greg"]}}],"url":["https://blog.example.com/progress-report#c1"]}},{"type":["https://schema.org/UserComments"],"properties":{"commentTime":["2013-08-29"],"creator":[{"type":["https://schema.org/Person"],"properties":{"name":["Charlotte"]}}],"url":["https://blog.example.com/progress-report#c2"]}}],"datePublished":["2013-08-29"],"headline":["Progress report"],"url":["https://blog.example.com/progress-report?comments=0"]}}]}`
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func BenchmarkParser(b *testing.B) {
	buf := bytes.NewBufferString(blogSnippet)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseHTML(buf, "utf-8", "https://blog.example.com/progress-report")
		if err != nil && err != io.EOF {
			b.Error(err)
		}
	}
}

func TestParseMicrodataPersonSnippet(t *testing.T) {
	buf := bytes.NewBufferString(personSnippet)
	data, err := ParseHTML(buf, "charset=utf-8", "https://jsonld.com/person/")
	if err != nil {
		t.Error(err)
	}

	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	result := string(b)
	expected := `{"items":[{"type":["Person"],"properties":{"@context":["https://schema.org"],"address":[{"type":["PostalAddress"],"properties":{"addressLocality":["Colorado Springs"],"addressRegion":["CO"],"postalCode":["80840"],"streetAddress":["100 Main Street"]}}],"alumniOf":["Dartmouth"],"birthDate":["1979-10-12"],"birthPlace":["Philadelphia, PA"],"colleague":["https://www.example.com/JohnColleague.html","https://www.example.com/JameColleague.html"],"email":["info@example.com"],"gender":["female"],"height":["72 inches"],"image":["janedoe.jpg"],"jobTitle":["Research Assistant"],"memberOf":["Republican Party"],"name":["Jane Doe"],"nationality":["Albanian"],"sameAs":["https://www.facebook.com/","https://www.linkedin.com/","https://twitter.com/","https://instagram.com/","https://plus.google.com/"],"telephone":["(123) 456-6789"],"url":["https://www.example.com"]}}]}`
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func TestParseMicrodataRecipeSnippet(t *testing.T) {
	buf := bytes.NewBufferString(recipeSnippet)
	data, err := ParseHTML(buf, "charset=utf-8", "https://jsonld.com/recipe/")
	if err != nil {
		t.Error(err)
	}

	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	result := string(b)
	expected := `{"items":[{"type":["Recipe"],"properties":{"@context":["https://schema.org"],"author":["Jake  Smith"],"cookTime":["PT2H"],"datePublished":["2015-05-18"],"description":["Your recipe description goes here"],"image":["https://www.example.com/images.jpg"],"interactionStatistic":[{"type":["InteractionCounter"],"properties":{"interactionType":["https://schema.org/Comment"],"userInteractionCount":["5"]}}],"name":["Rand's Cookies"],"nutrition":[{"type":["NutritionInformation"],"properties":{"calories":["1200 calories"],"carbohydrateContent":["12 carbs"],"fatContent":["9 grams fat"],"proteinContent":["9 grams of protein"]}}],"prepTime":["PT15M"],"recipeIngredient":["ingredient 1","ingredient 2","ingredient 3","ingredient 4","ingredient 5"],"recipeInstructions":["This is the long part, etc."],"recipeYield":["12 cookies"]}}]}`
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func TestParseMicrodataMultipleSnippet(t *testing.T) {
	buf := bytes.NewBufferString(multipleSnippet)
	data, err := ParseHTML(buf, "charset=utf-8", "https://www.w3.org/TR/json-ld11/")
	if err != nil {
		t.Error(err)
	}

	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	result := string(b)
	expected := `{"items":[{"type":["Person"],"properties":{"@context":["https://schema.org/"],"@id":["https://digitalbazaar.com/author/dlongley/"],"name":["Dave Longley"]}},{"type":["Person"],"properties":{"@context":["https://schema.org/"],"@id":["https://greggkellogg.net/foaf#me"],"name":["Gregg Kellogg"]}}]}`
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

// This HTML snippet is taken from the W3C Working Group website at https://html.spec.whatwg.org/multipage/microdata.html#global-identifiers-for-items
var bookSnippet = `
<dl itemscope
    itemtype="https://vocab.example.net/book"
    itemid="urn:isbn:0-330-34032-8">
 <dt>Title</dt>
 <dd itemprop="title">The Reality Dysfunction</dd>
 <dt>Author</dt>
 <dd itemprop="author">Peter F. Hamilton</dd>
 <dt>Publication date</dt>
 <dd><time itemprop="pubdate" datetime="1996-01-26">26 January 1996</time></dd>
</dl>`

// This HTML snippet is taken from the W3C Working Group website at https://html.spec.whatwg.org/multipage/microdata.html#associating-names-with-items
var gallerySnippet = `
<!DOCTYPE HTML>
<html>
 <head>
  <title>Photo gallery</title>
 </head>
 <body>
  <h1>My photos</h1>
  <figure itemscope itemtype="https://n.whatwg.org/work" itemref="licenses">
   <img itemprop="work" src="images/house.jpeg" alt="A white house, boarded up, sits in a forest.">
   <figcaption itemprop="title">The house I found.</figcaption>
  </figure>
  <figure itemscope itemtype="https://n.whatwg.org/work" itemref="licenses">
   <img itemprop="work" src="images/mailbox.jpeg" alt="Outside the house is a mailbox. It has a leaflet inside.">
   <figcaption itemprop="title">The mailbox.</figcaption>
  </figure>
  <footer>
   <p id="licenses">All images licensed under the <a itemprop="license"
   href="https://www.opensource.org/licenses/mit-license.php">MIT
   license</a>.</p>
  </footer>
 </body>
</html>`

// This HTML document is taken from the W3C Working Group website at https://html.spec.whatwg.org/multipage/microdata.html#json
var blogSnippet = `
<!DOCTYPE HTML>
<title>My Blog</title>
<article itemscope itemtype="https://schema.org/BlogPosting">
 <header>
  <h1 itemprop="headline">Progress report</h1>
  <p><time itemprop="datePublished" datetime="2013-08-29">today</time></p>
  <link itemprop="url" href="?comments=0">
 </header>
 <p>All in all, he's doing well with his swim lessons. The biggest thing was he had trouble
 putting his head in, but we got it down.</p>
 <section>
  <h1>Comments</h1>
  <article itemprop="comment" itemscope itemtype="https://schema.org/UserComments" id="c1">
   <link itemprop="url" href="#c1">
   <footer>
    <p>Posted by: <span itemprop="creator" itemscope itemtype="https://schema.org/Person">
     <span itemprop="name">Greg</span>
    </span></p>
    <p><time itemprop="commentTime" datetime="2013-08-29">15 minutes ago</time></p>
   </footer>
   <p>Ha!</p>
  </article>
  <article itemprop="comment" itemscope itemtype="https://schema.org/UserComments" id="c2">
   <link itemprop="url" href="#c2">
   <footer>
    <p>Posted by: <span itemprop="creator" itemscope itemtype="https://schema.org/Person">
     <span itemprop="name">Charlotte</span>
    </span></p>
    <p><time itemprop="commentTime" datetime="2013-08-29">5 minutes ago</time></p>
   </footer>
   <p>When you say "we got it down"...</p>
  </article>
 </section>
</article>`

// This HTML document is taken from the "Steal Our JSON-LD" website at https://jsonld.com/person/
var personSnippet = `<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@type": "Person",
  "address": {
	"@type": "PostalAddress",
	"addressLocality": "Colorado Springs",
	"addressRegion": "CO",
	"postalCode": "80840",
	"streetAddress": "100 Main Street"
  },
  "colleague": [
	"https://www.example.com/JohnColleague.html",
	"https://www.example.com/JameColleague.html"
  ],
  "email": "info@example.com",
  "image": "janedoe.jpg",
  "jobTitle": "Research Assistant",
  "name": "Jane Doe",
  "alumniOf": "Dartmouth",
  "birthPlace": "Philadelphia, PA",
  "birthDate": "1979-10-12",
  "height": "72 inches",
  "gender": "female",
  "memberOf": "Republican Party",
  "nationality": "Albanian",
  "telephone": "(123) 456-6789",
  "url": "https://www.example.com",
	"sameAs" : [ "https://www.facebook.com/",
  "https://www.linkedin.com/",
  "https://twitter.com/",
  "https://instagram.com/",
  "https://plus.google.com/"]
}
</script>`

// This HTML document is taken from the "Steal Our JSON-LD" website at https://jsonld.com/recipe/
var recipeSnippet = `<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@type": "Recipe",
  "author": "Jake  Smith",
  "cookTime": "PT2H",
  "datePublished": "2015-05-18",
  "description": "Your recipe description goes here",
  "image": "https://www.example.com/images.jpg",
  "recipeIngredient": [
	"ingredient 1",
	"ingredient 2",
	"ingredient 3",
	"ingredient 4",
	"ingredient 5"
  ],
  "interactionStatistic": {
	"@type": "InteractionCounter",
	"interactionType": "https://schema.org/Comment",
	"userInteractionCount": "5"
  },
  "name": "Rand's Cookies",
  "nutrition": {
	"@type": "NutritionInformation",
	"calories": "1200 calories",
	"carbohydrateContent": "12 carbs",
	"proteinContent": "9 grams of protein",
	"fatContent": "9 grams fat"
  },
  "prepTime": "PT15M",
  "recipeInstructions": "This is the long part, etc.",
  "recipeYield": "12 cookies"
}
</script>`

// This HTML document is taken from the "W3.org" website at https://www.w3.org/TR/json-ld11/
var multipleSnippet = `<p>Data describing Dave</p>
<script id="dave" type="application/ld+json">
{
  "@context": "https://schema.org/",
  "@id": "https://digitalbazaar.com/author/dlongley/",
  "@type": "Person",
  "name": "Dave Longley"
}
</script>

<p>Data describing Gregg</p>
<script id="gregg" type="application/ld+json">
{
  "@context": "https://schema.org/",
  "@id": "https://greggkellogg.net/foaf#me",
  "@type": "Person",
  "name": "Gregg Kellogg"
}
</script>`
