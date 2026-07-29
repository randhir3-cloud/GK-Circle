import { defineStore } from "pinia";
export const useUsersStore = defineStore(
  "users-store",
  () => {
    const userData = ref(null);

    const setUserData = (data) => {
      userData.value = data;
    };

    const getUserData = () => {
      return userData.value;
    };

    const fetchAuthenticatedUser = async () => {
      const { setUserDataStore } = await import("@/composables/auth");
      return await setUserDataStore();
    };

    return { userData, setUserData, getUserData, fetchAuthenticatedUser };
  },
  {
    persist: true,
  }
);
