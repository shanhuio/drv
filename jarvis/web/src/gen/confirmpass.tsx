import * as React from 'react' // for tsx

import go from '@shanhuio/htmlgen/dist/go'

import * as common from './common'

export function make() {
    let p = common.makePage('confirmpass')
    p.body = <div>
        <div id="main"></div>
        <script>var pageData={go('.Data')};</script>
    </div>

    p.scripts = [
        '/jslib/jquery.js',
        '/jslib/react.js',
        '/jslib/react-dom.js',
        '/js/confirmpass.js',
    ]
    return p
}
