import * as dashboard from '../dashboard'
import * as dashcore from '../dashcore'

declare var pageData: dashcore.PageData
jQuery(() => { dashboard.main(pageData) })
