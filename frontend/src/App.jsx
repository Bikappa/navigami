import QRCode from "qrcode"
import { useRef, useEffect, useState, useMemo, useCallback } from 'react'

function apiRequest(input, init) {
  return fetch(input, {
    headers: {
      'Accept': 'application/json',
    },
    ...init
  })
}
function useInterval(callback, delay) {
  const savedCallback = useRef();

  // Remember the latest callback.
  useEffect(() => {
    savedCallback.current = callback;
  }, [callback]);

  // Set up the interval.
  useEffect(() => {
    function tick() {
      savedCallback.current();
    }
    if (delay !== null) {
      let id = setInterval(tick, delay);
      return () => clearInterval(id);
    }
  }, [delay]);
}
function Portal() {
  const [sessionId, setSessionId] = useState()
  const [imageData, setImageData] = useState()
  useEffect(() => {
    if (sessionId) {
      return
    }

    apiRequest("/session".toString(), {
      method: "POST",
    }).then(res => {
      if (res.status !== 200) {
        throw new Error('Unexpected response code')
      }
      return res.text()
    }).then(setSessionId)

  }, [sessionId])

  const joinURL = useMemo(() => {
    return new URL(`/join?session=${sessionId}`, document.baseURI).href
  }, [sessionId])


  useEffect(() => {
    if (!joinURL) {
      return
    }
    console.log("ciao", { joinURL })

    QRCode.toDataURL(joinURL).then(setImageData)
  }, [sessionId])

  useInterval(() => {
    if (!sessionId) {
      return
    }
    apiRequest(`/session?session=${sessionId}`.toString()).then((res) => {
      if (!res.ok) {
        return
      }

      return res.text()

    }).then((gotoURL) => {
      if (!gotoURL) {
        return
      }
      window.location.href = gotoURL
    })
  }, 3000)

  if (!sessionId) {
    return <></>
  }

  return <div>
    <div>
      {imageData && <img src={imageData} />}
    </div>
    <div>
      <a href={joinURL} target="navigami-join">{sessionId.substring(0, 4)} - {sessionId.substring(4)} </a>
    </div>
  </div>
}


function Join() {
  const windowLocationHref = new URL(window.location.href)
  const sessionId = windowLocationHref.searchParams.get("session")
  const [gotoURL, setGotoUrl] = useState()
  const submitHandler = useCallback(async (e) => {
    e.preventDefault()
    return apiRequest(`/join?session=${sessionId}&gotoUrl=${gotoURL}`, {
      method: "POST",
    }).then(async res => {
      if (res.ok) {
        window.close()
        return
      }

      window.alert(await res.text())
    })
  })
  if (!sessionId) {
    return <>Invalid session id</>
  }

  console.log(gotoURL)
  return <>
    <form onSubmit={submitHandler}>
      <input onChange={(e) => {
        e.preventDefault()
        setGotoUrl(e.target.value)
      }} type="text" placeholder="type your url here" />
      <input type="submit" value="submit" />
    </form>
  </>
}

function App() {

  const windowLocationHref = new URL(window.location.href)
  return <>
    {windowLocationHref.pathname === '/join' ? <Join /> : <Portal />}
  </>
}

export default App
