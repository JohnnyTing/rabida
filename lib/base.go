package lib

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

// see: https://intoli.com/blog/not-possible-to-block-chrome-headless/
const script = `(function (w, n, wn) {

  Object.defineProperty(n, 'webdriver', {
    get: () => false,
  });

  function mockPluginsAndMimeTypes() {
    /* global MimeType MimeTypeArray PluginArray */

    // Disguise custom functions as being native
    const makeFnsNative = (fns = []) => {
      const oldCall = Function.prototype.call
      function call() {
        return oldCall.apply(this, arguments)
      }
      // eslint-disable-next-line
      Function.prototype.call = call

      const nativeToStringFunctionString = Error.toString().replace(
        /Error/g,
        'toString'
      )
      const oldToString = Function.prototype.toString

      function functionToString() {
        for (const fn of fns) {
          if (this === fn.ref) {
            return 'function ' + fn.name + '() { [native code] }'
          }
        }

        if (this === functionToString) {
          return nativeToStringFunctionString
        }
        return oldCall.call(oldToString, this)
      }
      // eslint-disable-next-line
      Function.prototype.toString = functionToString
    }

    const mockedFns = []

    const fakeData = {
      mimeTypes: [
        {
          type: 'application/pdf',
          suffixes: 'pdf',
          description: '',
          __pluginName: 'Chrome PDF Viewer'
        },
        {
          type: 'application/x-google-chrome-pdf',
          suffixes: 'pdf',
          description: 'Portable Document Format',
          __pluginName: 'Chrome PDF Plugin'
        },
        {
          type: 'application/x-nacl',
          suffixes: '',
          description: 'Native Client Executable',
          enabledPlugin: Plugin,
          __pluginName: 'Native Client'
        },
        {
          type: 'application/x-pnacl',
          suffixes: '',
          description: 'Portable Native Client Executable',
          __pluginName: 'Native Client'
        }
      ],
      plugins: [
        {
          name: 'Chrome PDF Plugin',
          filename: 'internal-pdf-viewer',
          description: 'Portable Document Format'
        },
        {
          name: 'Chrome PDF Viewer',
          filename: 'mhjfbmdgcfjbbpaeojofohoefgiehjai',
          description: ''
        },
        {
          name: 'Native Client',
          filename: 'internal-nacl-plugin',
          description: ''
        }
      ],
      fns: {
        namedItem: instanceName => {
          // Returns the Plugin/MimeType with the specified name.
          const fn = function (name) {
            if (!arguments.length) {
              throw new TypeError(
                'Failed to execute : 1 argument required, but only 0 present.'
              )
            }
            return this[name] || null
          }
          mockedFns.push({ ref: fn, name: 'namedItem' })
          return fn
        },
        item: instanceName => {
          // Returns the Plugin/MimeType at the specified index into the array.
          const fn = function (index) {
            if (!arguments.length) {
              throw new TypeError(
                'Failed to execute : 1 argument required, but only 0 present.'
              )
            }
            return this[index] || null
          }
          mockedFns.push({ ref: fn, name: 'item' })
          return fn
        },
        refresh: instanceName => {
          // Refreshes all plugins on the current page, optionally reloading documents.
          const fn = function () {
            return undefined
          }
          mockedFns.push({ ref: fn, name: 'refresh' })
          return fn
        }
      }
    }
    // Poor mans _.pluck
    const getSubset = (keys, obj) =>
      keys.reduce((a, c) => ({ ...a, [c]: obj[c] }), {})

    function generateMimeTypeArray() {
      const arr = fakeData.mimeTypes
        .map(obj => getSubset(['type', 'suffixes', 'description'], obj))
        .map(obj => Object.setPrototypeOf(obj, MimeType.prototype))
      arr.forEach(obj => {
        arr[obj.type] = obj
      })

      // Mock functions
      arr.namedItem = fakeData.fns.namedItem('MimeTypeArray')
      arr.item = fakeData.fns.item('MimeTypeArray')

      return Object.setPrototypeOf(arr, MimeTypeArray.prototype)
    }

    const mimeTypeArray = generateMimeTypeArray()
    Object.defineProperty(n, 'mimeTypes', {
      get: () => mimeTypeArray
    })

    function generatePluginArray() {
      const arr = fakeData.plugins
        .map(obj => getSubset(['name', 'filename', 'description'], obj))
        .map(obj => {
          const mimes = fakeData.mimeTypes.filter(
            m => m.__pluginName === obj.name
          )
          // Add mimetypes
          mimes.forEach((mime, index) => {
            n.mimeTypes[mime.type].enabledPlugin = obj
            obj[mime.type] = n.mimeTypes[mime.type]
            obj[index] = n.mimeTypes[mime.type]
          })
          obj.length = mimes.length
          return obj
        })
        .map(obj => {
          // Mock functions
          obj.namedItem = fakeData.fns.namedItem('Plugin')
          obj.item = fakeData.fns.item('Plugin')
          return obj
        })
        .map(obj => Object.setPrototypeOf(obj, Plugin.prototype))
      arr.forEach(obj => {
        arr[obj.name] = obj
      })

      // Mock functions
      arr.namedItem = fakeData.fns.namedItem('PluginArray')
      arr.item = fakeData.fns.item('PluginArray')
      arr.refresh = fakeData.fns.refresh('PluginArray')

      return Object.setPrototypeOf(arr, PluginArray.prototype)
    }

    const pluginArray = generatePluginArray()
    Object.defineProperty(n, 'plugins', {
      get: () => pluginArray
    })

    // Make mockedFns toString() representation resemble a native function
    makeFnsNative(mockedFns)
  }
  try {
    mockPluginsAndMimeTypes()
  } catch (err) { }

  // Pass the Languages Test.
  // Overwrite the plugins property to use a custom getter.
  Object.defineProperty(n, 'languages', {
    get: () => ["zh-CN", "und", "en", "zh-TW"],
  });

  // Pass the Languages Test.
  // Overwrite the plugins property to use a custom getter.
  Object.defineProperty(n, 'language', {
    get: () => "zh-CN",
  });


  function generateNetworkInformation() {
    let obj = {
      downlink: 4.8,
      effectiveType: "4g",
      onchange: null,
      rtt: 100,
      saveData: false
    }
    return Object.setPrototypeOf(obj, NetworkInformation.prototype)
  }

  const connection = generateNetworkInformation()
  Object.defineProperty(n, 'connection', {
    get: () => connection
  })

  w.outerHeight = 972
  w.outerWidth = 1920

  // Pass the Chrome Test.
  // We can mock this in as much depth as we need for the test.
  w.chrome = {
    app: {
      isInstalled: false,
    },
    webstore: {
      onInstallStageChanged: {},
      onDownloadProgress: {},
    },
    runtime: {
      PlatformOs: {
        MAC: 'mac',
        WIN: 'win',
        ANDROID: 'android',
        CROS: 'cros',
        LINUX: 'linux',
        OPENBSD: 'openbsd',
      },
      PlatformArch: {
        ARM: 'arm',
        X86_32: 'x86-32',
        X86_64: 'x86-64',
      },
      PlatformNaclArch: {
        ARM: 'arm',
        X86_32: 'x86-32',
        X86_64: 'x86-64',
      },
      RequestUpdateCheckStatus: {
        THROTTLED: 'throttled',
        NO_UPDATE: 'no_update',
        UPDATE_AVAILABLE: 'update_available',
      },
      OnInstalledReason: {
        INSTALL: 'install',
        UPDATE: 'update',
        CHROME_UPDATE: 'chrome_update',
        SHARED_MODULE_UPDATE: 'shared_module_update',
      },
      OnRestartRequiredReason: {
        APP_UPDATE: 'app_update',
        OS_UPDATE: 'os_update',
        PERIODIC: 'periodic',
      },
    },
  };

  ['height', 'width'].forEach(property => {
    // store the existing descriptor
    const imageDescriptor = Object.getOwnPropertyDescriptor(HTMLImageElement.prototype, property);

    // redefine the property with a patched descriptor
    Object.defineProperty(HTMLImageElement.prototype, property, {
      ...imageDescriptor,
      get: function () {
        // return an arbitrary non-zero dimension if the image failed to load
        if (this.complete && this.naturalHeight == 0) {
          return 16;
        }
        // otherwise, return the actual dimension
        return imageDescriptor.get.apply(this);
      },
    });
  })

  // Pass the Permissions Test.
  Object.defineProperty(Notification, 'permission', {
    get: () => 'default'
  });
  const originalQuery = wn.permissions.query;
  wn.permissions.query = (parameters) => (
    parameters.name === 'notifications' ?
      Promise.resolve({ state: 'prompt' }) :
      originalQuery(parameters)
  );

})(window, navigator, window.navigator);`

func evalJS(js string) chromedp.Tasks {
	var res *runtime.RemoteObject
	return chromedp.Tasks{
		chromedp.EvaluateAsDevTools(js, &res),
		chromedp.ActionFunc(func(ctx context.Context) error {
			b, err := res.MarshalJSON()
			if err != nil {
				return err
			}
			fmt.Println("result: ", string(b))
			return nil
		}),
	}
}

func waitLoaded(ctx context.Context) error {
	// TODO: this function is inherently racy, as we don't run ListenTarget
	// until after the navigate action is fired. For example, adding
	// time.Sleep(time.Second) at the top of this body makes most tests hang
	// forever, as they miss the load event.
	//
	// However, setting up the listener before firing the navigate action is
	// also racy, as we might get a load event from a previous navigate.
	//
	// For now, the second race seems much more common in real scenarios, so
	// keep the first approach. Is there a better way to deal with this?
	ch := make(chan bool)
	lctx, cancel := context.WithCancel(ctx)
	chromedp.ListenTarget(lctx, func(ev interface{}) {
		if _, ok := ev.(*page.EventLoadEventFired); ok {
			cancel()
			close(ch)
		}
	})
	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func Navigate(link string) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		_, _, _, err := page.Navigate(link).Do(ctx)
		if err != nil {
			return err
		}
		//time.Sleep(5)
		return waitLoaded(ctx)
	})
}
