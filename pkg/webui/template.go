package webui

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"unicode"
)

func renderForm(form string) []byte {
	return []byte(`
<html>
	<head>
		<title>DX LDAP</title>
		<style>
			* {
				margin: 0;
				padding: 0;
			}

			.container {
				width: 100%;
	  			height: 100%;
	  			position: relative;
			}

			.vertical-center {
  				position: absolute;
				left: 50%;
  				top: 50%;
  				-ms-transform: translateY(-50%) translateX(-50%);
  				transform: translateY(-50%) translateX(-50%);
			}
		</style>
		<script>
    		function escapeRegExp(str) {
      			return str.replace(/[\-\[\]\/\{\}\(\)\*\+\?\.\\\^\$\|]/g, "\\$&");
    		}
		</script>
	</head>
	<body>
		<div class="container">
  			<div class="vertical-center">
				<form id="form" method="POST">
`+form+`
				</form>
			</div>
		</div>
	</body>
</html>
`)
}

var findValuesExpr = regexp.MustCompile(`{([a-z]*)}`)
func replaceValues(in string, req *http.Request) string {
	matches := findValuesExpr.FindAllStringSubmatch(in, -1)
	for _, match := range matches {
		placeholder := match[0]
		fieldName := match[1]
		in = strings.ReplaceAll(in, placeholder, req.PostFormValue(fieldName))
	}
	return in
}

func validateUsername(username string) error {
	if len(username) == 0 {
		return fmt.Errorf("empty username")
	}
	for _, r := range username {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return fmt.Errorf("only [a-z0-9] is supported in an username")
		}
		if unicode.IsLetter(r) && !unicode.IsLower(r) {
			return fmt.Errorf("only *lowcased* characters and digits are supported in an username")
		}
	}
	return nil
}
