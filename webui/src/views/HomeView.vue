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
            
            pollInterval: null
        }
    },
    computed: {
        activeConversation() {
            return this.conversations.find(c => c.conversationId === this.activeConversationId);
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
                this.messages = response.data.reverse(); 
            } catch (e) {
                alert(e.toString());
            }
        },
        async sendMessage(content) {
            try {
                await this.$axios.post("/messages", {
                    conversationId: this.activeConversationId,
                    content: content,
                    contentType: "text"
                });
                await this.openConversation(this.activeConversationId);
                this.refreshConversations();
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
        async updateProfileName(newName) {
            try {
                await this.$axios.put("/user/name", { newName });
                this.username = newName;
                localStorage.setItem("username", this.username);
                alert("Name updated!");
            } catch (e) {
                alert(e.toString());
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
        initForward(msg) {
            this.messageToForward = msg;
            this.forwardTargets = [];
            this.showForwardModal = true;
        },
        async doForward() {
            if (this.forwardTargets.length === 0) return;
            try {
                await this.$axios.post("/messages/" + this.messageToForward.id + "/forward", {
                    targetConversationIds: this.forwardTargets
                });
                this.showForwardModal = false;
                alert("Forwarded!");
            } catch (e) {
                alert(e.toString());
            }
        },
        resolvePhotoUrl(url) {
            if (!url) return null;
            if (url.startsWith("http")) return url;
            if (url.startsWith("/")) {
                // If baseURL ends with / and url starts with /, strip one?
                // axios baseURL usually doesn't end with slash if configured as "http://localhost:3000".
                // But url is "/static/...".
                const base = this.$axios.defaults.baseURL || "";
                return base + url;
            }
            return url;
        },
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
                     this.messages = res.data.reverse();
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
             <div class="bg-dark rounded shadow p-3" style="width: 400px; max-height: 80vh; display: flex; flex-direction: column;">
                 <h5 class="text-white mb-3">Forward Message to...</h5>
                 <div class="flex-grow-1 overflow-auto custom-scrollbar mb-3 border border-secondary rounded">
                     <div v-for="c in conversations" :key="c.conversationId" class="p-2 border-bottom border-secondary d-flex align-items-center gap-2">
                         <input type="checkbox" :value="c.conversationId" v-model="forwardTargets" class="form-check-input bg-dark-input border-secondary">
                         <span class="text-white">{{ c.name || (c.isGroup ? 'Group ' + c.conversationId : 'Chat ' + c.conversationId) }}</span>
                     </div>
                 </div>
                 <div class="d-flex justify-content-end gap-2">
                     <button class="btn btn-secondary" @click="showForwardModal = false">Cancel</button>
                     <button class="btn btn-success" @click="doForward" :disabled="forwardTargets.length === 0">Forward</button>
                 </div>
             </div>
        </div>
    </div>
</template>

<style>
/* Global overrides/utilities if needed */
</style>
