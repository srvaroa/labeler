package labeler

import (
	"fmt"
	"log"
	"strings"

	gh "github.com/google/go-github/v35/github"
)

func NewAuthorCondition() Condition {
	return Condition{
		GetName: func() string {
			return "Author matches"
		},
		Evaluate: func(pr *gh.PullRequest, matcher LabelMatcher) (bool, error) {
			if len(matcher.Authors) <= 0 {
				return false, fmt.Errorf("Users are not set in config")
			}

			prAuthor := pr.GetUser().Login

			log.Printf("Matching `%s` against: `%v`", matcher.Authors, *prAuthor)
			for _, author := range matcher.Authors {
				if strings.ToLower(author) == strings.ToLower(*prAuthor) {
					return true, nil
				}
			}
			return false, nil
		},
	}
}
