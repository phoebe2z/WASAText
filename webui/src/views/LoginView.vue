<script>
export default {
	data: function() {
		return {
			errormsg: null,
			username: "",
		}
	},
	methods: {
		async login() {
			this.errormsg = null;
			if (this.username.length < 3) {
				this.errormsg = "Username must be at least 3 characters.";
				return;
			}
			try {
				let response = await this.$axios.post("/session", { name: this.username });
				// Save identifier (simple local storage or just state if not refreshing)
				// For homework, localStorage is good to persist login across refreshes.
				localStorage.setItem("user_id", response.data.identifier);
				localStorage.setItem("username", this.username);
				
				// Configure axios auth header globally or per request
				this.$axios.defaults.headers.common['Authorization'] = 'Bearer ' + response.data.identifier;
				
				this.$router.push("/chat");
			} catch (e) {
				if (e.response && e.response.status === 400) {
					this.errormsg = "Invalid username format/length.";
				} else {
					this.errormsg = e.toString();
				}
			}
		}
	},
	mounted() {
		if (localStorage.getItem("user_id")) {
			this.$axios.defaults.headers.common['Authorization'] = 'Bearer ' + localStorage.getItem("user_id");
			this.$router.push("/chat");
		}
	}
}
</script>

<template>
	<div class="d-flex justify-content-center align-items-center vh-100 bg-light">
		<div class="card p-4 shadow-sm" style="max-width: 400px; width: 100%;">
			<h2 class="text-center mb-4">WASAText</h2>
			<div class="mb-3">
				<label for="username" class="form-label">Username</label>
				<input type="text" class="form-control" id="username" v-model="username" @keyup.enter="login" placeholder="Enter your name">
			</div>
			<div class="d-grid">
				<button class="btn btn-primary" @click="login">Login / Register</button>
			</div>
			<div v-if="errormsg" class="alert alert-danger mt-3">
				{{ errormsg }}
			</div>
		</div>
	</div>
</template>

<style scoped>
</style>
