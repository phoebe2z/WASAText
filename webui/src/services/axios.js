import axios from "axios";

const baseURL = __API_URL__ || window.location.origin;

const instance = axios.create({
	baseURL: baseURL,
	timeout: 1000 * 5
});

export default instance;
