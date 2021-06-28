// Copyright (c) nano Authors. All Rights Reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package component

import (
	"github.com/aura-studio/nano/scheduler"
)

type (
	options struct {
		name          string                 // component name
		renameHandler func(string) string    // rename handler name
		schedule      scheduler.SchedFunc    // schedule service task
		dictionary    map[uint16]interface{} // Dictionary info slice
	}

	// Option used to customize handler
	Option func(options *options)
)

// WithName used to rename component name
func WithName(name string) Option {
	return func(opt *options) {
		opt.name = name
	}
}

// WithRenameHandlerFunc override handler name by specific function
// such as: strings.ToUpper/strings.ToLower
func WithRenameHandlerFunc(fn func(string) string) Option {
	return func(opt *options) {
		opt.renameHandler = fn
	}
}

// WithScheduleFunc set the func of the service schedule
func WithScheduleFunc(fn scheduler.SchedFunc) Option {
	return func(opt *options) {
		opt.schedule = fn
	}
}

// WithDictionary set dictionary for compressed route
func WithDictionary(dict map[uint16]interface{}) Option {
	return func(opt *options) {
		opt.dictionary = dict
	}
}
