import axios from "axios";
import MockAdapter from "axios-mock-adapter";

// Mock data
const axiosServices = axios.create({
  baseURL:
    process.env.NEXT_PUBLIC_REACT_APP_API_URL || "http://10.1.10.145:8888/",
});
var mock = new MockAdapter(axiosServices);

mock.onAny().passThrough();

export default axiosServices;
