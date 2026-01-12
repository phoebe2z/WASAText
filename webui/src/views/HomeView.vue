<script>
import SidebarRail from '../components/SidebarRail.vue';
import ChatList from '../components/ChatList.vue';
import ProfilePanel from '../components/ProfilePanel.vue';
import ChatWindow from '../components/ChatWindow.vue';
import RightPanel from '../components/RightPanel.vue';

export default {
    components: {
        SidebarRail, ChatList, ProfilePanel, ChatWindow, RightPanel
    },
    data: function() {
        return {
            loading: false,
            errormsg: null,
            userId: localStorage.getItem("user_id"),
            username: localStorage.getItem("username"),
            photoUrl: null,
            
            // Layout State
            activeTab: 'chat', // 'chat', 'profile'
            showRightPanel: false,
            
            // Data
            conversations: [],
            activeConversationId: null,
            messages: [],
            
            // Forward Modal
            showForwardModal: false,
            messageToForward: null,
            forwardTargets: [], // selected conversation Ids
            forwardUserTargets: [], // selected user IDs (for non-friends)
            forwardSearchQuery: "",
            allUsers: [],
            
            pollInterval: null
        }
    },
    computed: {
        activeConversation() {
            return this.conversations.find(c => c.conversationId === this.activeConversationId);
        },
        filteredForwardConversations() {
            if (!this.forwardSearchQuery) return this.conversations;
            const q = this.forwardSearchQuery.toLowerCase();
            return this.conversations.filter(c => (c.name || '').toLowerCase().includes(q));
        },
        filteredForwardUsers() {
            const q = this.forwardSearchQuery.toLowerCase();
            return this.allUsers.filter(u => 
                (u.name || '').toLowerCase().includes(q) && 
                u.id != this.userId &&
                !this.conversations.find(c => !c.isGroup && c.name === u.name)
            );
        }
    },
    methods: {
        async fetchUserProfile() {
            try {
                let response = await this.$axios.get("/user/me");
                this.username = response.data.name;
                this.photoUrl = this.resolvePhotoUrl(response.data.photoUrl);
                if (this.username) localStorage.setItem("username", this.username);
            } catch (e) {
                console.error("Error fetching profile:", e);
            }
        },
        async refreshConversations() {
            try {
                let response = await this.$axios.get("/conversations");
                this.conversations = response.data.map(c => ({
                    ...c,
                    photoUrl: this.resolvePhotoUrl(c.photoUrl)
                }));
            } catch (e) {
                console.error(e);
            }
        },
        async openConversation(id) {
            this.activeConversationId = id;
            this.messages = [];
            this.showRightPanel = false; // Close info on switch? or keep open? Usually close.
            try {
                let response = await this.$axios.get("/conversations/" + id);
                this.messages = response.data; 
            } catch (e) {
                alert(e.toString());
            }
        },
        async sendMessage(content, type = "text", replyToId = null) {
            try {
                let actualContent = content;
                let actualType = type;

                if (content instanceof File) {
                    actualContent = await this.fileToBase64(content);
                    actualType = "photo";
                }

                await this.$axios.post("/messages", {
                    conversationId: this.activeConversationId,
                    content: actualContent,
                    contentType: actualType,
                    replyToId: replyToId
                });
                await this.openConversation(this.activeConversationId);
                this.refreshConversations();
            } catch (e) {
                alert("Error sending message: " + (e.response ? (e.response.data || e.response.statusText) : e.message));
            }
        },
        fileToBase64(file) {
            return new Promise((resolve, reject) => {
                const reader = new FileReader();
                reader.readAsDataURL(file);
                reader.onload = () => resolve(reader.result);
                reader.onerror = error => reject(error);
            });
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
        async updateProfileName(newName) {
            try {
                await this.$axios.put("/user/name", { newName });
                this.username = newName;
                localStorage.setItem("username", this.username);
                alert("Name updated!");
            } catch (e) {
                if (e.response && e.response.status === 409) {
                    alert("Error: Username already taken.");
                } else {
                    alert(e.toString());
                }
            }
        },
        async updateProfilePhoto(payload) {
             try {
                let res;
                if (payload instanceof File) {
                    const formData = new FormData();
                    formData.append("newPhoto", payload);
                    res = await this.$axios.put("/user/photo", formData, {
                        headers: { 'Content-Type': 'multipart/form-data' }
                    });
                    alert("Photo updated via File Upload!");
                } else {
                    res = await this.$axios.put("/user/photo", { photoUrl: payload });
                    alert("Photo updated via URL!");
                }
                
                if (res && res.data && res.data.photoUrl) {
                    this.photoUrl = this.resolvePhotoUrl(res.data.photoUrl);
                }
             } catch(e) {
                 console.error(e);
                 alert("Error updating photo: " + e.toString());
             }
        },
        async createGroup() {
            const name = prompt("Enter Group Name:");
            if (!name) return;
            const membersStr = prompt("Enter Member IDs (comma separated):");
            if (!membersStr) return;
            
            let members = membersStr.split(",").map(id => parseInt(id.trim())).filter(id => !isNaN(id));
            if (members.length < 1) { 
                 alert("Need members");
                 return;
            }
             
            try {
                await this.$axios.post("/groups", {
                    name: name,
                    initialMembers: members
                });
                this.refreshConversations();
            } catch (e) {
                alert(e.toString());
            }
        },
        logout() {
            localStorage.removeItem("user_id");
            localStorage.removeItem("username");
            this.$router.push("/");
        },
        
        // --- Group Management ---
        onChatCreated(conversationId) {
            this.refreshConversations();
            this.openConversation(conversationId);
        },
        async setGroupName(groupId, newName) {
            try {
                await this.$axios.put("/groups/" + groupId + "/name", { newName });
                this.refreshConversations(); // Update list
                // Update active conversation object strictly? computed prop will pick up change from conversations list update
            } catch (e) {
                alert(e.toString());
            }
        },
        async setGroupPhoto(groupId, payload) {
             try {
                if (payload instanceof File) {
                    const formData = new FormData();
                    formData.append("photo", payload);
                    await this.$axios.put("/groups/" + groupId + "/photo", formData, {
                        headers: { 'Content-Type': 'multipart/form-data' }
                    });
                    this.refreshConversations();
                    alert("Group photo updated!");
                } else {
                    await this.$axios.put("/groups/" + groupId + "/photo", { photoUrl: payload });
                    this.refreshConversations();
                }
             } catch(e) {
                 console.error(e);
                 alert("Error updating group photo: " + e.toString());
             }
        },
        async addMember(groupId, userIds) {
            try {
                await this.$axios.post("/groups/" + groupId + "/members", { userIds });
                alert("Members added!");
            } catch (e) {
                alert(e.toString());
            }
        },
        async leaveGroup(groupId) {
             try {
                await this.$axios.delete("/groups/" + groupId + "/me");
                this.activeConversationId = null;
                this.refreshConversations();
            } catch (e) {
                alert(e.toString());
            }
        },
        
        // --- Reactions & Forwarding ---
        async reactMessage(msgId, emoticon) {
            try {
                await this.$axios.post("/messages/" + msgId + "/reaction", { emoticon }); // "Comment" = Reaction
                this.openConversation(this.activeConversationId); // Refresh to show
            } catch (e) {
               console.error(e);
            }
        },
        async unreactMessage(msgId, userId) {
            if (parseInt(this.userId) !== parseInt(userId)) return; // Can only remove own
            try {
                await this.$axios.delete("/messages/" + msgId + "/reaction");
                this.openConversation(this.activeConversationId);
            } catch (e) {
               console.error(e);
            }
        },
        async initForward(msg) {
            this.messageToForward = msg;
            this.forwardTargets = [];
            this.forwardUserTargets = [];
            this.forwardSearchQuery = "";
            this.showForwardModal = true;
            try {
                let res = await this.$axios.get("/users");
                this.allUsers = res.data;
            } catch (e) {
                console.error(e);
            }
        },
        async doForward() {
            if (this.forwardTargets.length === 0 && this.forwardUserTargets.length === 0) return;
            try {
                // Forwarding to conversations
                if (this.forwardTargets.length > 0) {
                    await this.$axios.post("/messages/" + this.messageToForward.id + "/forward", {
                        targetConversationIds: this.forwardTargets
                    });
                }
                
                // Forwarding to specific users (not yet in conversation)
                for (const uId of this.forwardUserTargets) {
                    const user = this.allUsers.find(u => u.id === uId);
                    if (user) {
                        let convId;
                        try {
                            let res = await this.$axios.post("/conversations", { recipientName: user.name });
                            convId = res.data.conversationId;
                        } catch (err) {
                            if (err.response && err.response.status === 400 && err.response.data.conversationId) {
                                convId = err.response.data.conversationId;
                            } else {
                                console.error("Error creating DM for forward", err);
                                continue;
                            }
                        }
                        
                        if (convId) {
                            await this.$axios.post("/messages/" + this.messageToForward.id + "/forward", {
                                targetConversationIds: [convId]
                            });
                        }
                    }
                }

                this.showForwardModal = false;
                alert("Forwarded!");
                this.refreshConversations();
            } catch (e) {
                alert(e.toString());
            }
        },
        resolvePhotoUrl(url) {
            if (!url) return null;
            if (url.startsWith("http")) return url;
            if (url.startsWith("/")) {
                // Use baseURL to derive absolute URL for images if needed
                const base = this.$axios.defaults.baseURL || "";
                return base + url;
            }
            return url;
        },
        handleTargetClick(type, id) {
            if (type === 'c') {
                const idx = this.forwardTargets.indexOf(id);
                if (idx > -1) this.forwardTargets.splice(idx, 1);
                else this.forwardTargets.push(id);
            } else {
                const idx = this.forwardUserTargets.indexOf(id);
                if (idx > -1) this.forwardUserTargets.splice(idx, 1);
                else this.forwardUserTargets.push(id);
            }
        }
    },
    mounted() {
        if (!this.userId) {
            this.$router.push("/");
            return;
        }
        this.$axios.defaults.headers.common['Authorization'] = 'Bearer ' + this.userId;
        this.refreshConversations();
        this.fetchUserProfile();
        
        this.pollInterval = setInterval(() => {
            this.refreshConversations();
            if (this.activeConversationId) {
                 this.$axios.get("/conversations/" + this.activeConversationId).then(res => {
                     this.messages = res.data;
                 }).catch(err => {
                     if (err.response && err.response.status === 404) {
                         // Selection is gone (e.g. left group or deleted)
                         this.activeConversationId = null;
                         this.messages = [];
                     }
                 });
            }
        }, 3000);
    },
    beforeUnmount() {
        if (this.pollInterval) clearInterval(this.pollInterval);
    }
}
</script>

<template>
    <div class="d-flex vh-100 w-100 overflow-hidden" style="background-color: #0c1317;">
        <!-- Left Rail -->
        <SidebarRail 
            v-model:activeTab="activeTab" 
        />
        
        <!-- Navigation Drawer -->
        <div class="d-flex flex-column h-100 border-end border-secondary" style="width: 400px; min-width: 300px;">
            <ChatList 
                v-if="activeTab === 'chat'" 
                :conversations="conversations" 
                :activeId="activeConversationId" 
                :currentUserId="userId"
                @select-chat="openConversation"
                @create-group="createGroup"
                @chat-created="onChatCreated"
            />
            <ProfilePanel 
                v-if="activeTab === 'profile'"
                :username="username"
                :userId="userId"
                :photoUrl="photoUrl"
                @update-name="updateProfileName"
                @update-photo="updateProfilePhoto"
                @logout="logout"
            />
        </div>
        
        <!-- Main Chat Window -->
        <div class="flex-grow-1 position-relative d-flex flex-column">
            <ChatWindow 
                v-if="activeConversationId && activeConversation"
                :conversation="activeConversation"
                :messages="messages"
                :currentUser="username"
                :userId="userId"
                @send-message="sendMessage"
                @delete-message="deleteMessage"
                @toggle-info="showRightPanel = !showRightPanel"
                @react-message="reactMessage"
                @unreact-message="unreactMessage"
                @forward-message="initForward"
            />
            
            <!-- Default Welcome State -->
            <div v-else class="h-100 d-flex flex-column justify-content-center align-items-center text-secondary border-bottom border-success border-5" style="background-color: #222e35;">
                <div class="mb-4">
                     <svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" stroke-linecap="round" stroke-linejoin="round" class="feather feather-monitor"><rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect><line x1="8" y1="21" x2="16" y2="21"></line><line x1="12" y1="17" x2="12" y2="21"></line></svg>
                </div>
                <h2 class="fw-light text-white mb-3">WASAText Web</h2>
                <div class="text-center small" style="max-width: 400px;">
                    Send and receive messages without keeping your phone online.<br>
                    Use WASAText on up to 4 linked devices and 1 phone.
                </div>
            </div>
        </div>
        
        <!-- Right Panel (Info) -->
         <RightPanel 
            v-if="showRightPanel && activeConversation"
            :conversation="activeConversation"
            :username="username"
            :userId="userId"
            @set-group-name="setGroupName"
            @set-group-photo="setGroupPhoto"
            @add-member="addMember"
            @leave-group="leaveGroup"
            @close="showRightPanel = false"
        />
        
        <!-- Forward Modal Overlay -->
        <div v-if="showForwardModal" class="position-absolute top-0 start-0 w-100 h-100 d-flex justify-content-center align-items-center bg-black bg-opacity-75" style="z-index: 2000;">
             <div class="bg-dark rounded-3 shadow-lg p-0 overflow-hidden" style="width: 400px; max-height: 80vh; display: flex; flex-direction: column; background-color: #111b21 !important;">
                 <!-- Header -->
                 <div class="p-3 d-flex align-items-center gap-3 text-white border-bottom border-secondary border-opacity-25" style="background-color: #202c33;">
                     <button class="btn btn-link text-white p-0" @click="showForwardModal = false">
                         <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-x"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
                     </button>
                     <h6 class="mb-0 fw-semibold">Forward message to</h6>
                 </div>
                 
                 <!-- Search -->
                 <div class="p-3">
                     <div class="input-group">
                         <span class="input-group-text bg-dark-input border-0 text-secondary pe-0">
                             <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-search"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
                         </span>
                         <input type="text" class="form-control bg-dark-input text-white border-0 forward-search-input" placeholder="Search name or number" v-model="forwardSearchQuery" style="box-shadow: none;">
                     </div>
                 </div>

                 <!-- List -->
                 <div class="flex-grow-1 overflow-auto custom-scrollbar px-1 pb-3">
                     <div v-if="filteredForwardConversations.length > 0 || filteredForwardUsers.length > 0">
                         <small class="text-secondary fw-semibold d-block px-3 py-2" style="font-size: 0.85rem;">Recent chats</small>
                         
                         <!-- Conversations -->
                         <div v-for="c in filteredForwardConversations" :key="'c'+c.conversationId" class="px-3 py-2 d-flex align-items-center gap-3 chat-item rounded-0 cursor-pointer" @click="handleTargetClick('c', c.conversationId)">
                             <input type="checkbox" :value="c.conversationId" v-model="forwardTargets" class="form-check-input bg-dark-input border-secondary m-0" @click.stop>
                             <div class="position-relative">
                                 <img v-if="c.photoUrl" :src="c.photoUrl" class="rounded-circle" style="width: 40px; height: 40px; object-fit: cover;">
                                 <div v-else class="rounded-circle d-flex align-items-center justify-content-center bg-secondary text-white" style="width: 40px; height: 40px;">
                                     <svg v-if="c.isGroup" xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-users"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="9" cy="7" r="4"></circle><path d="M23 21v-2a4 4 0 0 0-3-3.87"></path><path d="M16 3.13a4 4 0 0 1 0 7.75"></path></svg>
                                     <svg v-else xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-user"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle></svg>
                                 </div>
                             </div>
                             <div class="min-w-0 flex-grow-1 border-bottom border-secondary border-opacity-10 pb-2">
                                 <div class="text-white text-truncate">{{ c.name || (c.isGroup ? 'Group ' + c.conversationId : 'Chat ' + c.conversationId) }}</div>
                                 <small class="text-secondary d-block text-truncate">{{ c.isGroup ? 'Group' : 'Direct Message' }}</small>
                             </div>
                         </div>

                         <!-- Users (Friends/Searchable) -->
                         <div v-for="u in filteredForwardUsers" :key="'u'+u.id" class="px-3 py-2 d-flex align-items-center gap-3 chat-item rounded-0 cursor-pointer" @click="handleTargetClick('u', u.id)">
                             <input type="checkbox" :value="u.id" v-model="forwardUserTargets" class="form-check-input bg-dark-input border-secondary m-0" @click.stop>
                             <div class="position-relative">
                                 <img v-if="u.photoUrl" :src="resolvePhotoUrl(u.photoUrl)" class="rounded-circle" style="width: 40px; height: 40px; object-fit: cover;">
                                 <div v-else class="rounded-circle d-flex align-items-center justify-content-center bg-secondary text-white" style="width: 40px; height: 40px;">
                                     <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-user"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle></svg>
                                 </div>
                             </div>
                             <div class="min-w-0 flex-grow-1 border-bottom border-secondary border-opacity-10 pb-2">
                                 <div class="text-white text-truncate">{{ u.name }}</div>
                                 <small class="text-secondary d-block text-truncate">User</small>
                             </div>
                         </div>
                     </div>
                 </div>

                 <!-- Footer -->
                 <div class="p-3 d-flex justify-content-end bg-dark-header border-top border-secondary border-opacity-25 shadow-lg">
                     <button class="btn btn-success rounded-5 px-4 shadow-sm" @click="doForward" :disabled="forwardTargets.length === 0 && forwardUserTargets.length === 0" style="background-color: #00a884; border-color: #00a884;">
                         <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-send"><line x1="22" y1="2" x2="11" y2="13"></line><polygon points="22 2 15 22 11 13 2 9 22 2"></polygon></svg>
                     </button>
                 </div>
             </div>
        </div>
    </div>
</template>

<style>
.forward-search-input:focus {
    border: 1px solid #00a884 !important;
    background-color: #202c33 !important;
}
.chat-item:hover {
    background-color: #202c33;
}
</style>
