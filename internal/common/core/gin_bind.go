package core

import (
	"encoding/json"
	"go-file-server/internal/common/global"
	"io"
	"net/http"

	"github.com/gin-gonic/gin/binding"
	"github.com/pkg/errors"

	"github.com/gin-gonic/gin"
)

var (
	BindUri   = uriBinding{}
	BindJson  = jsonBinding{}
	BindForm  = formBinding{}
	BindQuery = queryBinding{}
)

type Bindings interface {
	Bind(c *gin.Context, obj any) error
}

// BindUri
type uriBinding struct{}

func (u uriBinding) Bind(c *gin.Context, obj any) error {
	m := make(map[string][]string, len(c.Params))
	for _, v := range c.Params {
		m[v.Key] = []string{v.Value}
	}
	return binding.MapFormWithTag(obj, m, "uri")
}

// BindForm
const defaultMemory = 32 << 20

type formBinding struct{}

func (formBinding) Bind(c *gin.Context, obj any) error {
	req := c.Request
	if err := req.ParseForm(); err != nil {
		return err
	}
	if err := req.ParseMultipartForm(defaultMemory); err != nil && !errors.Is(err, http.ErrNotMultipart) {
		return err
	}
	return binding.MapFormWithTag(obj, req.Form, "form")
}

// BindQuery
type queryBinding struct{}

func (queryBinding) Bind(c *gin.Context, obj any) error {
	values := c.Request.URL.Query()
	return binding.MapFormWithTag(obj, values, "form")
}

// BindJson
type jsonBinding struct{}

func (jsonBinding) Bind(c *gin.Context, obj any) error {
	if c.Request == nil || c.Request.Body == nil {
		return errors.Errorf("invalid request")
	}
	return decodeJSON(c.Request.Body, obj)
}

func decodeJSON(r io.Reader, obj any) error {
	decoder := json.NewDecoder(r)
	if binding.EnableDecoderUseNumber {
		decoder.UseNumber()
	}
	if binding.EnableDecoderDisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return nil
}

func ShouldBinds(c *gin.Context, obj any, bindings ...Bindings) error {
	for _, b := range bindings {
		if err := b.Bind(c, obj); err != nil {
			return NewApiErr(err).SetHttpCode(global.BadRequestError)
		}
	}
	return binding.Validator.ValidateStruct(obj)
}
