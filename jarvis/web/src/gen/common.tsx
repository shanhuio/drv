import * as page from '@shanhuio/htmlgen/dist/page'

export function makePage(name: string) {
    let prop = {
        title: 'HomeDrive Admin',
        css: '/style.css',
    }
    return new page.Page(name, prop)
}
