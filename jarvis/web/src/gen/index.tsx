import * as writer from '@shanhuio/htmlgen/dist/writer'

import * as cover from './cover'
import * as dashboard from './dashboard'
import * as confirmpass from './confirmpass'
import * as inputtotp from './inputtotp'

export function generate(dir: string) {
    writer.writePages(dir, [
        cover.make(),
        dashboard.make(),
        confirmpass.make(),
        inputtotp.make()
    ])
}

generate('tmpl')
