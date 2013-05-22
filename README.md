excerpt
=======

excerpt is a library written in go to search for weighted terms in a body of text and return scored excerpts containing those terms.

Example:
--------

```
example_text := "The the the the in this text. We want to find the excerpt of this text that contains the search_words."
search_ex := map[string]float64{"Excerpt": 1, "the": 0.05}
all_excerpts := FindExcerpts(search_ex, example_text, 20, false)
fmt.Println(all_excerpts)
```

finds all possible excerpts and scores them:

```
[<ExcerptWindow(charlength: 20|bytes: 20): starts at: 0 score: 0.6000000000000001>:
The the the the in t
 <ExcerptWindow(charlength: 19|bytes: 20): starts at: 4 score: 0.45000000000000007>:
the the the in this
 <ExcerptWindow(charlength: 20|bytes: 20): starts at: 8 score: 0.30000000000000004>:
the the in this text
 <ExcerptWindow(charlength: 20|bytes: 20): starts at: 12 score: 0.15000000000000002>:
the in this text. We
 <ExcerptWindow(charlength: 19|bytes: 20): starts at: 46 score: 7.15>:
the excerpt of this
 <ExcerptWindow(charlength: 20|bytes: 20): starts at: 50 score: 7>:
excerpt of this text
 <ExcerptWindow(charlength: 17|bytes: 17): starts at: 85 score: 0.15000000000000002>:
the search_words.
]
```

TODO: As you can see the algorithm is not too smart: From the match positions and string length it could be infered that matches 1,2,3 would all be scored lower than 0...

If you're only interested in the best match you can set a flag.
Multibyte characters are fine too:

```
example_text_jp := "日本語とか中国語でも大丈夫です。１バイト以上のunicodeの記号でもちゃんと出来ます。日本語が大丈夫。"
search_ex_jp := map[string]float64{"日本語": 1, "大丈夫": 1}
best_excerpts := FindExcerpts(search_ex_jp, example_text_jp, 8, true)[0]
fmt.Println(best_excerpts)

```

will get you:

```
<ExcerptWindow(charlength: 8|bytes: 24): starts at: 118 score: 6>:
日本語が大丈夫。
```

FindExcerpts searches searchterms in body and returns excerpts of a given
length that contain the terms.

The terms can be weighted and each ExcerptWindow gets a cummulative score
of the searchterms it contains. setting findHighestScore the function will
only return the ExcerptWindow with the hightest score.

If a match is at the end of a window and overlaps its boundry the window
will be extended to include the full match. Trailing whitespace is removed
so that you could get an excerpt that is shorter then the specified length.

An ExcerptWindow always starts with a match. In the future an option might
be added to position/center the window around the matches.
