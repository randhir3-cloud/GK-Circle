import { useUsersStore } from "~~/store/users";

export const returnToPathFromUrl = (absoluteUrl) => {
  if (!absoluteUrl) {
    return "";
  }
  try {
    const url = new URL(absoluteUrl);
    if (url.origin !== window.location.origin || url.pathname === "/") {
      return "";
    }
    return url.pathname + url.search;
  } catch (error) {
    console.log("unusable return_to on the kratos flow", error.message);
    return "";
  }
};

export const handleLogout = async () => {
  const { setUserData } = useUsersStore();
  const { kratosUrl } = useRuntimeConfig().public;
  try {
    // Step 1: Fetch logout URL and token from the first API endpoint
    const response = await fetch(`${kratosUrl}/self-service/logout/browser`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
      },
      credentials: "include",
    });

    if (!response.ok) {
      throw new Error("Failed to fetch logout URL and token");
    }

    const { logout_url } = await response.json();

    // Step 2: Use the fetched URL and token to make another API call
    const secondApiResponse = await fetch(`${logout_url}`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
      },
      credentials: "include",
    });

    if (!secondApiResponse.ok) {
      throw new Error(
        "Failed to perform logout with the provided URL and token"
      );
    }
    console.log("Logged out successfully");
    setUserData(null);
    navigateTo("/");
  } catch (error) {
    console.error("Error during logout:", error);
  }
};

const normalizeAuthenticatedUser = (data) => ({
  role: data?.role,
  avatar: data?.avatar,
  firstname: data?.firstname,
  username: data?.username,
  email: data?.email,
  canCreatePublicQuiz: !!data?.can_create_public_quiz,
});

export const getAuthenticatedUser = async () => {
  const { apiUrl } = useRuntimeConfig().public;
  const headers = useRequestHeaders(["cookie"]);
  const requestOptions = {
    method: "GET",
    credentials: "include",
    headers,
    mode: "cors",
  };
  const fetchWho = () => fetch(apiUrl + "/user/who", requestOptions);

  try {
    const sessionResponse = await fetch(
      apiUrl + "/kratos/whoami",
      requestOptions
    );
    if (!sessionResponse.ok) {
      return null;
    }

    let response = await fetchWho();

    if (response.status == 404) {
      await fetch(apiUrl + "/kratos/auth", {
        method: "GET",
        credentials: "include",
        headers: headers,
        mode: "cors",
        redirect: "manual",
      }).catch((error) => {
        console.log("unable to sync the kratos identity", error.message);
      });
      response = await fetchWho();
    }

    if (!response.ok) return null;

    const data = await response.json();
    return data?.data ? normalizeAuthenticatedUser(data.data) : null;
  } catch (error) {
    console.warn("Authentication status could not be verified.", {
      cause: error?.name || "request_failed",
    });
    return null;
  }
};

export const setUserDataStore = async () => {
  const { setUserData } = useUsersStore();
  const user = await getAuthenticatedUser();
  setUserData(user);
  return user;
};
