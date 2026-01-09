<script>
export default {
	data: function() {
		return {
			loading: false,
			errormsg: null,
			userId: localStorage.getItem("user_id"),
			username: localStorage.getItem("username"),
			conversations: [],
			activeConversationId: null,
			messages: [],
			newMessage: "",
			
			// Modal states (simplified toggles)
			showCreateGroup: false,
			showProfile: false,
			
			// Profile inputs
			newParamName: "",
			
			// Group inputs
			newGroupName: "",
			newGroupMembers: "", // comma separated IDs for simplicity or separate input
		}
	},
	computed: {
		activeConversation() {
			return this.conversations.find(c => c.conversationId === this.activeConversationId);
		}
	},
	methods: {
		async refreshConversations() {
			try {
				let response = await this.$axios.get("/conversations");
				this.conversations = response.data;
				if (this.conversations.length > 0 && !this.activeConversationId) {
					// Don't auto-select to avoid unread clearing or just select first
					// this.openConversation(this.conversations[0].conversationId);
				}
			} catch (e) {
				this.errormsg = e.toString();
			}
		},
		async openConversation(id) {
			this.activeConversationId = id;
			this.messages = []; // Clear old while loading
			try {
				let response = await this.$axios.get("/conversations/" + id);
				this.messages = response.data.reverse(); // API returns reverse chron (newest first)? Messages usually displayed oldest at top. 
				// Wait, "reverse chronologically" implies [Newest, ..., Oldest].
				// Chat UI typically renders [Oldest, ..., Newest].
				// So I need to reverse it to display naturally (top to bottom).
				this.scrollToBottom();
			} catch (e) {
				this.errormsg = e.toString();
			}
		},
		async sendMessage() {
			if (!this.newMessage.trim() || !this.activeConversationId) return;
			try {
				let payload = {
					conversationId: this.activeConversationId,
					content: this.newMessage,
					contentType: "text"
				};
				await this.$axios.post("/messages", payload);
				this.newMessage = "";
				await this.openConversation(this.activeConversationId); // Refresh to see new message
			} catch (e) {
				this.errormsg = e.toString();
			}
		},
		async createGroup() {
			let members = this.newGroupMembers.split(",").map(id => parseInt(id.trim())).filter(id => !isNaN(id));
			if (members.length < 2) {
				alert("Need at least 2 other members");
				return;
			}
			try {
				await this.$axios.post("/groups", {
					name: this.newGroupName,
					initialMembers: members
				});
				this.newGroupName = "";
				this.newGroupMembers = "";
				this.showCreateGroup = false;
				this.refreshConversations();
			} catch (e) {
				alert(e.toString());
			}
		},
		async updateProfileName() {
			try {
				await this.$axios.put("/user/name", { newName: this.newParamName });
				this.username = this.newParamName;
				localStorage.setItem("username", this.username);
				this.showProfile = false;
				alert("Name updated!");
			} catch (e) {
				alert(e.toString());
			}
		},
		async deleteMessage(id) {
			if (!confirm("Delete message?")) return;
			try {
				await this.$axios.delete("/messages/" + id);
				this.openConversation(this.activeConversationId);
			} catch (e) {
				alert(e.toString());
			}
		},
		logout() {
			localStorage.removeItem("user_id");
			localStorage.removeItem("username");
			this.$router.push("/");
		},
		scrollToBottom() {
			this.$nextTick(() => {
				const container = this.$refs.msgContainer;
				if (container) container.scrollTop = container.scrollHeight;
			});
		},
		formatTime(t) {
			return new Date(t).toLocaleTimeString();
		}
	},
	mounted() {
		if (!this.userId) {
			this.$router.push("/");
			return;
		}
		// Set auth header again just in case (if refresh happened)
		this.$axios.defaults.headers.common['Authorization'] = 'Bearer ' + this.userId;
		this.refreshConversations();
		
		// Poll for new messages every 5s
		setInterval(() => {
			if (this.activeConversationId) {
				// Silent refresh logic could be better (check for new only), but this works for homework
				// this.openConversation(this.activeConversationId); 
				// Polling might reset scroll, careful.
			}
			this.refreshConversations();
		}, 5000);
	}
}
</script>

<template>
	<div class="container-fluid vh-100 d-flex flex-column p-0">
		<!-- Header -->
		<header class="navbar navbar-dark bg-dark flex-md-nowrap p-2 shadow">
			<span class="navbar-brand col-md-3 col-lg-2 me-0 px-3 fs-6">WASAText - {{ username }}</span>
			<div class="d-flex gap-2">
				<button class="btn btn-sm btn-outline-light" @click="showProfile = !showProfile">Profile</button>
				<button class="btn btn-sm btn-outline-light" @click="showCreateGroup = !showCreateGroup">New Group</button>
				<button class="btn btn-sm btn-outline-danger" @click="logout">Logout</button>
			</div>
		</header>
		
		<!-- Modals/Overlays (Simplified inline) -->
		<div v-if="showProfile" class="alert alert-secondary m-2">
			<h5>Profile Settings</h5>
			<input v-model="newParamName" placeholder="New Name" class="form-control mb-2">
			<button class="btn btn-primary btn-sm" @click="updateProfileName">Update Name</button>
			<button class="btn btn-secondary btn-sm" @click="showProfile=false">Close</button>
		</div>

		<div v-if="showCreateGroup" class="alert alert-info m-2">
			<h5>Create Group</h5>
			<input v-model="newGroupName" placeholder="Group Name" class="form-control mb-2">
			<input v-model="newGroupMembers" placeholder="Member IDs (comma separated)" class="form-control mb-2">
			<button class="btn btn-success btn-sm" @click="createGroup">Create</button>
			<button class="btn btn-secondary btn-sm" @click="showCreateGroup=false">Close</button>
		</div>

		<div class="row flex-grow-1 g-0" style="overflow: hidden;">
			<!-- Sidebar -->
			<nav class="col-md-3 col-lg-2 bg-light border-end d-flex flex-column" style="overflow-y: auto;">
				<div class="list-group list-group-flush">
					<button 
						v-for="c in conversations" 
						:key="c.conversationId"
						class="list-group-item list-group-item-action"
						:class="{ active: activeConversationId === c.conversationId }"
						@click="openConversation(c.conversationId)"
					>
						<div class="d-flex w-100 justify-content-between">
							<h6 class="mb-1">{{ c.name || (c.isGroup ? 'Group ' + c.conversationId : 'Chat ' + c.conversationId) }}</h6>
							<small>{{ c.unreadCount > 0 ? '('+c.unreadCount+')' : '' }}</small>
						</div>
						<small class="text-muted text-truncate d-block">{{ c.latestMessagePreview }}</small>
					</button>
				</div>
			</nav>

			<!-- Chat Area -->
			<main class="col-md-9 col-lg-10 d-flex flex-column h-100">
				<div v-if="activeConversationId" class="flex-grow-1 p-3 d-flex flex-column" style="overflow: hidden;">
					<!-- Messages List -->
					<div class="flex-grow-1 overflow-auto mb-3" ref="msgContainer">
						<div v-for="msg in messages" :key="msg.id" class="mb-2">
							<div class="d-flex flex-column" :class="{ 'align-items-end': msg.senderName === username, 'align-items-start': msg.senderName !== username }">
								<div class="card p-2" :class="{ 'bg-primary text-white': msg.senderName === username, 'bg-light': msg.senderName !== username }" style="max-width: 75%;">
									<small class="text-white-50 mb-1" v-if="msg.senderName === username">You</small>
									<small class="text-muted mb-1" v-else>{{ msg.senderName }}</small>
									
									<div>{{ msg.content }}</div>
									
									<div class="mt-1 d-flex gap-2">
										<small style="font-size: 0.7em;">{{ formatTime(msg.timeStamp) }}</small>
										<button v-if="msg.senderName === username" class="btn btn-link btn-sm p-0 text-white" style="font-size: 0.7em;" @click="deleteMessage(msg.id)">Delete</button>
									</div>
								</div>
							</div>
						</div>
					</div>
					
					<!-- Composer -->
					<div class="input-group">
						<input type="text" class="form-control" v-model="newMessage" @keyup.enter="sendMessage" placeholder="Type a message...">
						<button class="btn btn-primary" type="button" @click="sendMessage">Send</button>
					</div>
				</div>
				
				<div v-else class="d-flex justify-content-center align-items-center h-100">
					<p class="text-muted">Select a conversation to start chatting</p>
				</div>
			</main>
		</div>
	</div>
</template>

<style scoped>
</style>
