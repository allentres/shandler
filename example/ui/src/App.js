import React, {useState, useReducer} from "react"
import "./styles.scss"

import {useWS} from './WS'

const initialState = {
	messages: [],
}
const stateReducer = (state, msg) => {
	let newState = {...state}
	let newMessages = [...state.messages, msg]
	if (newMessages.length > 5) {
		newMessages = newMessages.slice(-5)
	}
	newState = {...newState, messages: newMessages}
	return newState
}

export default function App() {
	const [state, setState] = useReducer(stateReducer, initialState)
	const [ws, wsStatus] = useWS(`ws://localhost:8080/ws`, msg => setState(msg))
	const sendData = () => {
		ws.sendJson({"a": "b", "c": 3})
	}

	return <div className={`App`}>
		<a href={`/ws.client.html`}>WS Client {wsStatus}</a>
		<br/>
		<button onClick={sendData}>Send</button>
		<br/>
		{state.messages.map((msg, key) => <div className={`code`} key={key}>{JSON.stringify(msg)}</div>)}
	</div>
}