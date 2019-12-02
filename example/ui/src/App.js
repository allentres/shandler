import React, {useState} from "react"
import "./styles.scss"

export default function App() {
	const [hidden, setHidden] = useState(true)
	let content
	if (hidden) {
		content = <button onClick={() => setHidden(!hidden)}>Click Me!</button>
	} else {
		content = <div className={`code`} onClick={() => setHidden(!hidden)}>Hot React Hooks App</div>
	}
	return <div className={`App`}>{content}</div>
}