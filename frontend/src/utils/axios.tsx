import axios from "axios";
import MockAdapter from "axios-mock-adapter";

// Mock data
const axiosServices = axios.create({
  baseURL:
    process.env.NEXT_PUBLIC_REACT_APP_API_URL || "http://127.0.0.1:8888/",
});

axiosServices.interceptors.request.use(
  (request) => {
    console.error(request);
    return request;
  },
  (error) => {
    console.error(error);
  }
);

var mock = new MockAdapter(axiosServices);

mock.onAny().passThrough();

export default axiosServices;
