import React from "react";
import ReactDOM from "react-dom";
import {AppContainer} from "react-hot-loader"
import App from "./App"
import "./index.html"
import "./favicon.ico"
import "./ws.client.html"

const r = C => ReactDOM.render(<AppContainer><C/></AppContainer>, document.querySelector("#root"))
r(App)
if (module.hot) module.hot.accept(`./App`, () => r(App))
