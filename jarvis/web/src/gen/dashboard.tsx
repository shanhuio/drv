import * as React from 'react' // for tsx

import go from '@shanhuio/htmlgen/dist/go'

import * as common from './common'

export function make() {
    let body = <div>
        <div className="main" id="main"></div>
        <script>var pageData={go('.Data')};</script>
    </div>

    let p = common.makePage('dashboard')
    p.body = body
    p.bodyClass = 'dashboard'
    p.title = 'HomeDrive'
    p.scripts = [
        '/jslib/jquery.js',
        '/jslib/react.js',
        '/jslib/react-dom.js',
        '/js/dashboard.js',
    ]
    return p
}
