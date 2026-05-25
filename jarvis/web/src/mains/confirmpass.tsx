import * as confirmpass from '../confirmpass'

declare var pageData: confirmpass.PageData
jQuery(() => { confirmpass.main(pageData) })
