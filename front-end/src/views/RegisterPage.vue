<script setup lang="ts">
import { ref, onMounted, nextTick } from "vue";
import { useRouter } from "vue-router";
import { usePhoto } from "@/composables/usePhoto";

const { uploadPhoto } = usePhoto();

const router = useRouter();

const currentStep = ref(1);

const userEmail = ref("");

const form = ref({
  username: "",
  name: "",
  email: "",
  password: "",
  confpassword: "",
  gender: "",
  dob: "",
  newsletter_subscribed: false,
  turnstile_token: "",
  two_fa_enabled: false,
  profile_picture: "",
});

const otpForm = ref({
  otp: "",
  email: "",

});

const loading = ref(false);
const error = ref("");
const success = ref("");
const authToken = ref("");
const turnstileLoaded = ref(false);
const turnstileWidgetId = ref<string | null>(null);
const rateLimitSeconds = ref(0);
const rateLimitInterval = ref<number | null>(null);

const errors = ref({
  username: "",
  name: "",
  email: "",
  password: "",
  confpassword: "",
  gender: "",
  dob: "",
});

onMounted(async () => {
  await nextTick();

  if (document.querySelector("script[src*=\"turnstile\"]")) {
    turnstileLoaded.value = true;
    renderTurnstile();
    return;
  }

  const script = document.createElement("script");
  script.src = "https://challenges.cloudflare.com/turnstile/v0/api.js";
  script.async = true;
  script.defer = true;
  script.onload = () => {
    turnstileLoaded.value = true;
    renderTurnstile();
  };
  document.head.appendChild(script);
});


const renderTurnstile = async () => {
  await nextTick();
  const container = document.querySelector("#turnstile-container");

  if (!container || !window.turnstile || turnstileWidgetId.value) {
    return;
  }

  if (container.innerHTML.trim() !== "") {
    return;
  }

  try {
    turnstileWidgetId.value = window.turnstile.render("#turnstile-container", {
      sitekey: import.meta.env.VITE_TURNSTILE_API_KEY,
      callback: (token: string) => {
        form.value.turnstile_token = token;
      },
      "error-callback": () => {
        error.value = "Turnstile verification failed. Please refresh the page.";
      },
    });
  } catch (err) {
    console.error("Turnstile render error:", err);
  }
};

const otpCountdown = (seconds: number) => {
  if (rateLimitInterval.value) {
    clearInterval(rateLimitInterval.value);
  }

  rateLimitSeconds.value = seconds;

  rateLimitInterval.value = setInterval(() => {
    rateLimitSeconds.value--;

    if (rateLimitSeconds.value <= 0) {
      if (rateLimitInterval.value) {
        clearInterval(rateLimitInterval.value);
        rateLimitInterval.value = null;
      }
    }
  }, 1000) as unknown as number;
};

const validateEmail = (email: string): boolean => {
  if (!email) return false;
  const regex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.com$/;
  return regex.test(email);
};

const validateName = (name: string): boolean => {
  if (!name) return false;
  const regex = /^[a-zA-Z\s]+$/;
  return regex.test(name);
};

const validatePassword = (password: string): string[] => {
  const errors: string[] = [];
  if (!password) return errors;
  if (password.length < 8) errors.push("Password must be at least 8 characters");
  if (!/[A-Z]/.test(password)) errors.push("Password must contain at least one uppercase letter");
  if (!/[0-9]/.test(password)) errors.push("Password must contain at least one number");
  if (!/[!@#$%^&*(),.?":{}|<>]/.test(password))
    errors.push("Password must contain at least one symbol");
  return errors;
};

const validateAge = (dob: string): boolean => {
  const birthDate = new Date(dob);
  const today = new Date();
  const age = today.getFullYear() - birthDate.getFullYear();
  const monthDiff = today.getMonth() - birthDate.getMonth();

  if (monthDiff < 0 || (monthDiff === 0 && today.getDate() < birthDate.getDate())) {
    return age - 1 >= 13;
  }
  return age >= 13;
};

const validateField = (field: string) => {
  switch (field) {
    case "email":
      if (!validateEmail(form.value.email)) {
        errors.value.email = "Email must end with .com domain";
      } else {
        errors.value.email = "";
      }
      break;
    case "username":
      if (form.value.username == "") {
        errors.value.username = "Username is required";
      } else {
        errors.value.username = "";
      }
      break;
    case "name":
      if (form.value.name == "") {
        errors.value.name = "Name is required";
      } else {
        errors.value.name = "";
      }
      if (!validateName(form.value.name)) {
        errors.value.name = "Name can only contain letters and spaces";
      } else {
        errors.value.name = "";
      }
      break;
    case "password":
      const pwdErrors = validatePassword(form.value.password);
      errors.value.password = pwdErrors[0] || "";
      break;

    case "confpassword":
      if (form.value.password !== form.value.confpassword) {
        errors.value.confpassword = "Passwords do not match";
      } else {
        errors.value.confpassword = "";
      }
      break;

    case "gender":
      if (form.value.gender != "Male" && form.value.gender != "Female") {
        errors.value.gender = "Gender must be either male or female";
      }
      break;
    case "dateOfBirth":
      if (form.value.dob == "") {
        errors.value.dob = "Date of Birth is required";
      } else {
        errors.value.dob = "";
      }
      if (!validateAge(form.value.dob)) {
        errors.value.dob = "You must be at least 13 years old";
      } else {
        errors.value.dob = "";
      }
      break;
  }
};

const submit = async () => {
  error.value = "";
  success.value = "";
  validateField("username");
  validateField("name");

  validateField("email");
  validateField("password");
  validateField("confpassword");
  validateField("dateOfBirth");

  if (Object.values(errors.value).some(err => err !== "")) {
    error.value = "Please fix all validation errors";
    return;
  }

  loading.value = true;

  try {
    const response = await fetch("https://localhost/api/v1/users/register", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(form.value),
    });

    const data = await response.json();

    if (response.ok) {
      const registeredEmail = form.value.email;
      userEmail.value = registeredEmail;
      otpForm.value.email = registeredEmail;

      if (data.message && data.message.includes("OTP")) {
        success.value = "Registration successful! OTP has been sent to your email.";
      } else {
        success.value = "Registration successful!";
      }

      setTimeout(() => {
        currentStep.value = 2;
        error.value = "";
        success.value = "OTP has been sent to your email. Please check your inbox.";
        otpCountdown(60);
      }, 1500);
    } else {
      error.value = data.error || "Registration failed. Please try again.";

      if (window.turnstile && turnstileWidgetId.value) {
        window.turnstile.reset(turnstileWidgetId.value);
      }
    }
  } catch {
    error.value = "Network error. Please check your connection and try again.";

    if (window.turnstile && turnstileWidgetId.value) {
      window.turnstile.reset(turnstileWidgetId.value);
    }
  } finally {
    loading.value = false;
  }
};

const verifyOTP = async () => {
  error.value = "";
  success.value = "";


  if (!otpForm.value.otp || otpForm.value.otp.length !== 6) {
    error.value = "Please enter a valid 6-digit OTP code";
    return;
  }

  loading.value = true;

  try {
    const requestBody = {
      email: otpForm.value.email,
      code: otpForm.value.otp,
    };

    const otpResponse = await fetch("https://localhost/api/v1/users/verify-otp", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(requestBody),
    });

    const data = await otpResponse.json();

    if (otpResponse.ok) {
      if (data.token) {
        authToken.value = data.token;
        localStorage.setItem("token", data.token);
      }

      success.value = "Email verified successfully!";
      if (!form.value.turnstile_token) {
        error.value = "Please complete the CAPTCHA verification";
        return;
      }
      setTimeout(() => {
        currentStep.value = 3;
        error.value = "";
        success.value = "";
      }, 1500);
    } else {
      error.value = data.error || "Invalid OTP code. Please try again.";
    }
  } catch (err) {
    error.value = "Network error. Please try again.";
    console.error("OTP verification error:", err);
  } finally {
    loading.value = false;
  }
};

const selectedFile = ref<File | null>(null);
const uploadedPhotoUrl = ref("");

const handleFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement;
  if (target.files && target.files[0]) {
    selectedFile.value = target.files[0];

    const allowedTypes = ["image/jpeg", "image/jpg", "image/png"];
    if (!allowedTypes.includes(target.files[0].type)) {
      error.value = "Please select a valid image file (JPEG or PNG)";
      selectedFile.value = null;
      return;
    }

    if (target.files[0].size > 5 * 1024 * 1024) {
      error.value = "File size must be less than 5MB";
      selectedFile.value = null;
      return;
    }

    error.value = "";
  }
};

const uploadPhotoHere = async () => {
  if (!selectedFile.value) {
    error.value = "Please select a photo to upload";
    return false;
  }

  const url = await uploadPhoto(selectedFile.value);
  if (!url) {
    error.value = "Photo upload failed. Please try again.";
    return false;
  }

  uploadedPhotoUrl.value = url;
  form.value.profile_picture = url;
  return true;
};

const completion = async () => {
  loading.value = true;
  error.value = "";

  try {
    if (selectedFile.value) {
      const uploadSuccess = await uploadPhotoHere();
      if (!uploadSuccess) {
        loading.value = false;
        return;
      }
      success.value = "Photo uploaded successfully!";
    }

    const response = await fetch("https://localhost/api/v1/users/complete-registration", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        email: userEmail.value,
        name: form.value.name,
        profile_picture: uploadedPhotoUrl.value || form.value.profile_picture,
        newsletter_subscribed: form.value.newsletter_subscribed,
        two_fa_enabled: form.value.two_fa_enabled,
      }),
    });

    const data = await response.json();

    if (response.ok) {
      success.value = "Registration complete! Welcome email sent.";

      setTimeout(() => {
        router.push("/login");
      }, 2000);
    } else {
      error.value = data.error || "Failed to complete registration";
    }
  } catch (err) {
    error.value = "Network error. Please try again.";
    console.error("Complete registration error:", err);
  } finally {
    loading.value = false;
  }
};

const resendOTP = async () => {
  if (rateLimitSeconds.value > 0) {
    return;
  }
  error.value = "";
  success.value = "";
  loading.value = true;

  const emailToSend = otpForm.value.email || userEmail.value;

  if (!emailToSend) {
    error.value = "Email address not found. Please try again.";
    loading.value = false;
    return;
  }

  try {
    const response = await fetch("https://localhost/api/v1/users/send-otp", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ email: emailToSend }),
    });

    const data = await response.json();

    if (response.ok) {
      success.value = "OTP code has been resent to your email!";
      rateLimitSeconds.value = 60;
      otpCountdown(60);
    } else {
      const errorMsg = data.error || "";
      const match = errorMsg.match(/wait (\d+) seconds/);

      if (match && match[1]) {
        const seconds = parseInt(match[1]);
        otpCountdown(seconds);
        error.value = `Please wait ${seconds} seconds before requesting a new OTP`;
      } else {
        error.value = errorMsg || "Failed to resend OTP. Please try again.";
      }
    }
  } catch (err) {
    error.value = "Network error. Please try again.";
    console.error("Resend OTP error:", err);
  } finally {
    loading.value = false;
  }
};

const handleGoogleLogin = () => {
  window.location.href = "https://localhost/api/v1/users/auth/google";
};

declare global {
  interface Window {
    turnstile: {
      render: (container: string, options: any) => string
      reset: (widgetId: string) => void
    }
  }
}
</script>
<template>
  <div class="register-wrapper">
    <div class="register-container">
      <div class="register-card">
        <img
          src="../components/icons/logo-dark.png"
          alt="Logo"
          width="100%"
          height="100"
        />

        <div class="step-indicator">
          <div
            class="step"
            :class="{ active: currentStep === 1, completed: currentStep > 1 }"
          />
          <div
            class="step-line"
            :class="{ active: currentStep > 1 }"
          />
          <div
            class="step"
            :class="{ active: currentStep === 2, completed: currentStep > 2 }"
          />
          <div
            class="step-line"
            :class="{ active: currentStep > 2 }"
          />
          <div
            class="step"
            :class="{ active: currentStep === 3 }"
          />
        </div>

        <div v-if="currentStep === 1">
          <p class="subtitle">
            <b style="font-weight: 600">Sign up to see photos and videos from your friends</b>
          </p>
          <br />
          <form
            id="register-form"
            @submit.prevent="submit"
          >
            <div class="form-group">
              <input
                id="username"
                v-model="form.username"
                type="text"
                required
                placeholder="Username"
                :disabled="loading"
              />
              <br />
              <span
                v-if="errors.username"
                class="error-text"
              >{{ errors.username }}</span>
            </div>

            <div class="form-group">
              <input
                id="name"
                v-model="form.name"
                type="text"
                required
                placeholder="Name"
                :disabled="loading"
              />
              <br />
              <span
                v-if="errors.name"
                class="error-text"
              >{{ errors.name }}</span>
            </div>

            <div class="form-group">
              <input
                id="email"
                v-model="form.email"
                type="email"
                required
                placeholder="Email"
                :disabled="loading"
                @blur="validateField('email')"
              />
              <br />
              <span
                v-if="errors.email"
                class="error-text"
              >{{ errors.email }}</span>
            </div>

            <div class="form-group">
              <input
                id="password"
                v-model="form.password"
                type="password"
                required
                placeholder="Password"
                :disabled="loading"
                @blur="validateField('password')"
              />
              <br />
              <span
                v-if="errors.password"
                class="error-text"
              >{{ errors.password }}</span>
            </div>

            <div class="form-group">
              <input
                id="confpassword"
                v-model="form.confpassword"
                type="password"
                required
                placeholder="Confirm password"
                :disabled="loading"
                @blur="validateField('confpassword')"
              />
              <br />
              <span
                v-if="errors.confpassword"
                class="error-text"
              >{{ errors.confpassword }}</span>
            </div>

            <div class="form-group">
              <select
                id="gender"
                v-model="form.gender"
                required
                :disabled="loading"
              >
                <option
                  value=""
                  disabled
                >
                  Select gender
                </option>
                <option value="Male">
                  Male
                </option>
                <option value="Female">
                  Female
                </option>
              </select>
            </div>

            <div class="form-group">
              <label for="dateOfBirth">Date of Birth *</label>
              <input
                id="dateOfBirth"
                v-model="form.dob"
                type="date"
                required
                :disabled="loading"
                @blur="validateField('dateOfBirth')"
              />
              <br />
              <span
                v-if="errors.dob"
                class="error-text"
              >{{ errors.dob }}</span>
            </div>

            <div class="form-group">
              <div id="turnstile-container" />
            </div>

            <div
              v-if="error"
              class="alert alert-error"
              style="color: red"
            >
              {{ error }}
            </div>
            <div
              v-if="success"
              class="alert alert-success"
              style="color: greenyellow"
            >
              {{ success }}
            </div>
            <button
              type="submit"
              class="btn-primary"
              :disabled="loading"
            >
              Register
            </button>

            <div class="divider">
              <span>OR</span>
            </div>

            <button
              type="button"
              class="btn-google"
              :disabled="loading"
              @click="handleGoogleLogin"
            >
              Continue with Google
            </button>

            <p
              class="footer-text"
              style="font-size: 12px"
            >
              Already have an account?
              <router-link
                to="/login"
                style="text-decoration: none; background-color: transparent; font-weight: 500"
              >
                Login here
              </router-link>
            </p>
          </form>
        </div>

        <div
          v-if="currentStep === 2"
          class="otp-section"
        >
          <h2>Verify Your Email</h2>
          <p class="otp-info">
            We've sent a 6-digit code to <strong>{{ userEmail }}</strong>
          </p>

          <form @submit.prevent="verifyOTP">
            <div class="form-group">
              <input
                v-model="otpForm.otp"
                type="text"
                maxlength="6"
                pattern="[0-9]{6}"
                required
                placeholder="Enter 6 digit OTP"
                :disabled="loading"
                class="otp-input"
              />
            </div>

            <div
              v-if="error"
              class="error-message"
            >
              {{ error }}
            </div>
            <div
              v-if="success"
              class="success-message"
            >
              {{ success }}
            </div>

            <button
              type="submit"
              class="btn-primary"
              :disabled="loading || otpForm.otp.length !== 6"
            >
              {{ loading ? 'Verifying...' : 'Verify OTP' }}
            </button>
          </form>

          <div class="otp-actions">
            <button  
              class="btn-link" 
              :disabled="loading || rateLimitSeconds > 0"
              @click="resendOTP"
            >
              {{ rateLimitSeconds > 0 ? `Resend OTP (${rateLimitSeconds}s)` : 'Resend OTP' }}
            </button>
            <p
              class="footer-text"
              style="font-size: 12px"
            >
              Already have an account?
              <router-link
                to="/login"
                style="text-decoration: none; background-color: transparent; font-weight: 500"
              >
                Login here
              </router-link>
            </p>
          </div>
        </div>

        <div
          v-if="currentStep === 3"
          class="completion-section"
        >
          <h2>Registration Complete!</h2>
          <div class="form-group">
            <label
              for="profile_picture"
              class="file-label"
            >
              {{ selectedFile ? selectedFile.name : 'Choose Profile Picture' }}
            </label>
            <input
              id="profile_picture"
              type="file"
              accept="image/*"
              :disabled="loading"
              class="file-input"
              @change="handleFileSelect"
            />
            <p
              v-if="selectedFile"
              class="file-info"
            >
              {{ (selectedFile.size / 1024).toFixed(2) }} KB
            </p>
          </div>

          <div class="form-group checkbox-group">
            <label>
              <input
                v-model="form.newsletter_subscribed"
                type="checkbox"
                :disabled="loading"
              />
              Subscribe to newsletter
            </label>
          </div>

          <div class="form-group checkbox-group">
            <label>
              <input
                v-model="form.two_fa_enabled"
                type="checkbox"
                :disabled="loading"
              />
              Enable Two-Factor Authentication
            </label>
          </div>

          <button 
            :disabled="loading"
            class="btn-primary"  
            @click="completion"
          >
            Continue To Login
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
* {
  max-height: 100vh;
  color: white;
  background-color: black;
}

.register-wrapper {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  width: 100%;
  background-color: black;
}

.register-container {
  border: 1px solid gainsboro;
  padding: 20px;
  border-radius: 20px;
  width: 30vw;
  display: flex;
  flex-direction: column;
  justify-content: center;
  text-align: center;
  align-items: center;
}

.form-group input {
  border: white 1px solid;
  border-radius: 10px;
  width: 75%;
  padding: 10px;
  background-color: black;
  color: white;
  height: 30px;
}

.form-group {
  margin-bottom: 15px;
}

.error-text {
  font-weight: 500;
  color: red;
  font-size: small;
}

button {
  width: 80%;
  height: 30px;
  border-radius: 5px;
  font-weight: 700;
  border: none;
  cursor: pointer;
  background-color: blueviolet;
}

.checkbox-group {
  display: flex;
  justify-content: center;
  align-items: center;
}

.checkbox-group label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.checkbox-group input[type='checkbox'] {
  width: auto;

  margin: 0;
  cursor: pointer;
}

.step-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 30px 0;
  gap: 10px;
}

.step {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  border: 2px solid #555;
  background-color: black;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  transition: all 0.3s ease;
}

.step.active {
  border-color: blueviolet;
  background-color: blueviolet;
}

.step.completed {
  border-color: green;
  background-color: green;
}

.step-line {
  width: 80px;
  height: 2px;
  background-color: #555;
  transition: all 0.3s ease;
}

.step-line.active {
  background-color: blueviolet;
}

.otp-section {
  text-align: center;
}

.otp-info {
  margin: 15px 0;
  color: #ccc;
}

.otp-input {
  font-size: 20px;

  text-align: center;
}

.otp-actions {
  margin-top: 20px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.btn-link {
  background: none;
  color: blueviolet;
  text-decoration: underline;
  border: none;
  cursor: pointer;
  width: auto;
}

.btn-link:hover {
  color: #a855f7;
}

.error-message {
  color: red;
  margin: 10px 0;
}

.success-message {
  color: greenyellow;
  margin: 10px 0;
  font-weight: 500;
}

.file-input {
  display: none;
}

.file-label {
  display: inline-block;
  padding: 12px 24px;
  background-color: #333;
  border: 2px dashed #666;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.3s ease;
  text-align: center;
  width: 75%;
  margin: 10px auto;
}

.file-label:hover {
  background-color: #444;
  border-color: blueviolet;
}

.file-info {
  font-size: 12px;
  color: #999;
  margin-top: 5px;
}
</style>
