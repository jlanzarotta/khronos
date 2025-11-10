/*
Copyright © 2018-2025 Jeff Lanzarotta
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

 1. Redistributions of source code must retain the above copyright notice,
    this list of conditions and the following disclaimer.

 2. Redistributions in binary form must reproduce the above copyright notice,
    this list of conditions and the following disclaimer in the documentation
    and/or other materials provided with the distribution.

 3. Neither the name of the copyright holder nor the names of its contributors
    may be used to endorse or promote products derived from this software
    without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.
*/

package rest

import (
	"encoding/base64"
	"khronos/constants"
	"khronos/internal/models"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/fatih/color"
	"github.com/spf13/viper"
)

// HTTPClient is a default HTTP client, a proxy over http.Client.
var HTTPClient Client

// Client represents and interface for http.Client. Its main purpose is testing.
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

func init() {
	HTTPClient = &http.Client{Timeout: time.Second * 30}
}

func ReadCredentials() models.Credentials {
	var username string = viper.GetString(constants.PUSH_USERNAME)
	if stringUtils.IsEmpty(username) {
		log.Fatalf("%s: Missing push username.  Please correct your configuration.\n", color.RedString(constants.FATAL_NORMAL_CASE))
		os.Exit(1)
	}

	var apiKey string = viper.GetString(constants.PUSH_API_KEY)
	if stringUtils.IsEmpty(apiKey) {
		log.Fatalf("%s: Missing push api_key.  Please correct your configuration.\n", color.RedString(constants.FATAL_NORMAL_CASE))
		os.Exit(1)
	}

	var credentials models.Credentials
	credentials.Username = username
	credentials.Password = apiKey

	return credentials
}

func BasicAuth(cred *models.Credentials) string {
	auth := cred.Username + ":" + cred.Password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
