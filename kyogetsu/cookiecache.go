/* Copyright Dylan Enloe 2017
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package kyogetsu

import (
  "net/http"
)

//A CookiesCache represents an interface to a data
//store for the purpose of storing and retrieving
//http.Cookie data
type CookieCache interface {
  //Stores a single http.Cookie.
  //id is the unique identifier for the session
  SetCookie(id string, c *http.Cookie) error
  //Stores an array of http.Cookies
  //id is the unique identifier for the session
  SetCookies(id string, c []*http.Cookie) error

  //Get cookie data for the given key and user session
  GetCookie(id string, key string) (*http.Cookie, error)
  //Gets all cookie data for a given user session
  GetCookies(id string) ([]*http.Cookie, error)

  //Change the Id that the cookie data is stored under
  ChangeCookiesId(old_id string, new_id string) error
}

