import React, {useState, useEffect, useRef} from 'react'

const MAX_MISSED_HEARTBEATS = 3
const PING_DELAY_MS = 20000
const RECONNECT_DELAY_MS = 1000
export const useWS = (wsHost, handleMessage) => {
	const [status, setStatus] = useState(`disconnected`)
	const heartBeatInterval = useRef(undefined)
	const missedHeartBeats = useRef(0)
	let [ws, setWS] = useState(false)

	const connectToServer = () => {
		setStatus(`connecting`)
		let _ws = new WebSocket(wsHost)
		_ws.sendJson = data => _ws.send(JSON.stringify(data))
		console.log(`[WS] --> ${wsHost}`)
		_ws.onopen = () => {
			console.info(`[WS] connected`)
			setStatus(`connected`)
			setWS(_ws)
			missedHeartBeats.current = 0
			if (heartBeatInterval.current) clearInterval(heartBeatInterval.current)
			heartBeatInterval.current = setInterval(() => {
					console.debug(`[WS] missed heartbeats`, missedHeartBeats.current)
					try {
						if (missedHeartBeats.current >= MAX_MISSED_HEARTBEATS) {
							console.error(`[WS] missed more than ${MAX_MISSED_HEARTBEATS} heartbeats, closing...`)
							_ws.close()
						} else {
							_ws.send(JSON.stringify({re: `ping`}))
							missedHeartBeats.current = missedHeartBeats.current + 1
						}
					} catch (e) {
					}
				}, PING_DELAY_MS
			)
		}
		_ws.onclose = () => {
			setStatus(`disconnected`)
			setWS(false)
			if (heartBeatInterval.current) clearInterval(heartBeatInterval.current)
			heartBeatInterval.current = undefined
			setTimeout(connectToServer, RECONNECT_DELAY_MS)
		}

		_ws.onerror = (_e) => {
			setStatus(`disconnected`)
			setWS(false)
		}
		_ws.onmessage = payload => {
			try {
				let msg = JSON.parse(payload.data)
				if (msg.re === `pong`) {
					missedHeartBeats.current = missedHeartBeats - 1
					if (missedHeartBeats.current > 0) {
						console.warn(`[WS] missed ${missedHeartBeats.current} (${MAX_MISSED_HEARTBEATS} max) heartbeats.`)
					}
				} else {
					handleMessage(msg)
				}
			} catch (e) {
				// most probably not a json payload, send it as is
				handleMessage(payload.data)
			}
		}
	}

	useEffect(() => {
		connectToServer()
		return () => {
			if (ws) ws.close()
		}
	}, [])
	return [ws, status]
}