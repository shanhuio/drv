import * as inputtotp from '../inputtotp'

declare var pageData: inputtotp.PageData
jQuery(() => { inputtotp.main(pageData) })
