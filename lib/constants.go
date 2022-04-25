package lib

const (
	DebugPrint = `console.log(
			"navigator.platform:", navigator.platform,
			"navigator.userAgent:", navigator.userAgent,
			"navigator.webdriver:", navigator.webdriver,
			"navigator.plugins.length:", navigator.plugins.length,
			"navigator.language:", navigator.language,
			"navigator.oscpu:", navigator.oscpu,
			"navigator.productSub:", navigator.productSub,
			"eval.toString().length:", eval.toString().length,
			"navigator.hardwareConcurrency:", navigator.hardwareConcurrency,
			"window.sessionStorage:", !!window.sessionStorage,
			"window.localStorage:", !!window.localStorage,
			"window.indexedDB:", !!window.indexedDB,
			"window.openDatabase:", !!window.openDatabase,
			"window.screen.width:", window.screen.width,
			"window.screen.availWidth:", window.screen.availWidth,
			"window.screen.height:", window.screen.height,
			"window.screen.availHeight:", window.screen.availHeight,
			"window.hasLiedResolution:", window.screen.width < window.screen.availWidth || window.screen.height < window.screen.availHeight,
		)`
)
