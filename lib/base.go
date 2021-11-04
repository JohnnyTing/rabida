package lib

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"math/rand"
	"time"
)

// Script see: https://intoli.com/blog/not-possible-to-block-chrome-headless/
const Script = `(function(w, n, wn) {
  // Pass the Webdriver Test.
  Object.defineProperty(n, 'webdriver', {
    get: () => false,
  });

  // Pass the Plugins Length Test.
  // Overwrite the plugins property to use a custom getter.
  Object.defineProperty(n, 'plugins', {
    // This just needs to have length > 0 for the current test,
    // but we could mock the plugins too if necessary.
    get: () => [1, 2, 3, 4, 5],
  });

  // Pass the Languages Test.
  // Overwrite the plugins property to use a custom getter.
  Object.defineProperty(n, 'languages', {
    get: () => ['en-US', 'en'],
  });

  // Pass the Chrome Test.
  // We can mock this in as much depth as we need for the test.
  w.chrome = {
    runtime: {},
  };

  // Pass the Permissions Test.
  const originalQuery = wn.permissions.query;
  return wn.permissions.query = (parameters) => (
    parameters.name === 'notifications' ?
      Promise.resolve({ state: Notification.permission }) :
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

func RandDuration(min, max time.Duration) time.Duration {
	from := min.Milliseconds()
	to := max.Milliseconds()
	rand.Seed(time.Now().Local().UnixNano())
	return time.Millisecond * time.Duration(rand.Int63n(to-from)+from)
}
