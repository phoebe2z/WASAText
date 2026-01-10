<script>
export default {
    props: ['conversation', 'username', 'userId'],
    emits: ['set-group-name', 'set-group-photo', 'add-member', 'leave-group', 'close'],
    data() {
        return {
            isEditingName: false,
            editNameValue: this.conversation.name || "",
            newMemberId: ""
        }
    },
    watch: {
        conversation: {
            handler(val) {
                this.editNameValue = val.name || "";
            },
            deep: true
        }
    },
    methods: {
        saveName() {
            if (this.editNameValue !== this.conversation.name) {
                this.$emit('set-group-name', this.conversation.conversationId, this.editNameValue);
            }
            this.isEditingName = false;
        },
        openPhotoUpload() {
             const url = prompt("Enter new Group Photo URL:");
             if (url) {
                 this.$emit('set-group-photo', this.conversation.conversationId, url);
             }
        },
        addMember() {
            if (!this.newMemberId) return;
            // Parse comma separated if needed, but UI says "Member ID". 
            // The API takes a list.
            const ids = this.newMemberId.split(",").map(id => parseInt(id.trim())).filter(id => !isNaN(id));
            if (ids.length > 0) {
                this.$emit('add-member', this.conversation.conversationId, ids);
                this.newMemberId = "";
            }
        },
        leaveGroup() {
            if (confirm("Are you sure you want to leave this group?")) {
                this.$emit('leave-group', this.conversation.conversationId);
            }
        }
    }
}
</script>

<template>
    <div class="d-flex flex-column h-100 bg-dark-list border-start border-secondary" style="width: 300px;">
        <!-- Header -->
        <div class="p-3 bg-dark-header d-flex align-items-center">
            <button class="btn btn-link text-secondary p-0 me-3" @click="$emit('close')">
                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-x"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
            </button>
            <h5 class="m-0 text-white">{{ conversation.isGroup ? 'Group Info' : 'Contact Info' }}</h5>
        </div>

        <div class="flex-grow-1 p-3 overflow-auto custom-scrollbar">
            <!-- Photo -->
            <div class="d-flex flex-column align-items-center mb-4 mt-3">
                <div class="position-relative mb-3" style="width: 200px; height: 200px; cursor: pointer;" @click="conversation.isGroup ? openPhotoUpload() : null" :title="conversation.isGroup ? 'Change Group Photo' : ''">
                    <div class="w-100 h-100 rounded-circle bg-secondary d-flex align-items-center justify-content-center overflow-hidden">
                        <img v-if="conversation.photoUrl" :src="conversation.photoUrl" class="w-100 h-100" style="object-fit: cover;">
                        <span v-else class="text-white fs-1">{{ (conversation.name || 'C').charAt(0).toUpperCase() }}</span>
                    </div>
                     <div v-if="conversation.isGroup" class="position-absolute top-0 start-0 w-100 h-100 rounded-circle d-flex align-items-center justify-content-center bg-dark bg-opacity-50 text-white opacity-0 hover-opacity-100 transition">
                        <div class="text-center small">Change<br>Photo</div>
                    </div>
                </div>
                
                <!-- Name -->
                <div class="w-100 text-center">
                    <div v-if="!isEditingName" class="d-flex justify-content-center align-items-center gap-2">
                        <h4 class="text-white m-0 text-break">{{ conversation.name || (conversation.isGroup ? 'Group ' + conversation.conversationId : 'User ' + conversation.conversationId) }}</h4>
                        <button v-if="conversation.isGroup" class="btn btn-link text-secondary p-0" @click="isEditingName = true">
                            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-edit-2"><path d="M17 3a2.828 2.828 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z"></path></svg>
                        </button>
                    </div>
                     <div v-else class="d-flex gap-2 justify-content-center">
                        <input v-model="editNameValue" class="form-control form-control-sm bg-dark-input text-white border-bottom border-0 rounded-0 p-0 text-center" @keyup.enter="saveName">
                        <button class="btn btn-link text-secondary p-0" @click="saveName">
                            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-check"><polyline points="20 6 9 17 4 12"></polyline></svg>
                        </button>
                    </div>
                     <small class="text-secondary" v-if="conversation.isGroup">Group • {{ conversation.conversationId }}</small>
                </div>
            </div>
            
            <div v-if="conversation.isGroup">
                 <hr class="border-secondary">
                 
                 <!-- Add Member -->
                 <div class="mb-4">
                     <label class="text-success small mb-2">Add Participants</label>
                     <div class="input-group input-group-sm">
                         <input v-model="newMemberId" placeholder="User ID" class="form-control bg-dark-input text-white border-secondary">
                         <button class="btn btn-success" @click="addMember">Add</button>
                     </div>
                 </div>
                 
                 <hr class="border-secondary">
                 
                 <!-- Leave Group -->
                 <button class="btn btn-outline-danger w-100 d-flex align-items-center justify-content-center gap-2" @click="leaveGroup">
                     <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-log-out"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path><polyline points="16 17 21 12 16 7"></polyline><line x1="21" y1="12" x2="9" y2="12"></line></svg>
                     Exit Group
                 </button>
            </div>
        </div>
    </div>
</template>

<style scoped>
.bg-dark-list { background-color: #111b21; }
.bg-dark-header { background-color: #202c33; }
.bg-dark-input { background-color: transparent; border-color: #2a3942; }
.bg-dark-input:focus { background-color: #2a3942; box-shadow: none; color: white; }

.text-secondary { color: #8696a0 !important; }
.text-success { color: #00a884 !important; }

.hover-opacity-100:hover { opacity: 1 !important; }
.transition { transition: opacity 0.2s; }
</style>
