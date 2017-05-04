package kyogetsu

type CookieCache interface {
  SetCookie(id string, key string, val string) error
  SetCookies(id string, val map[string]string) error

  GetCookie(id string, key string) (string, error)
  GetCookies(id string) (map[string]string, error)
}

