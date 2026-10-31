// Connects layer 1 to the records: every packet that arrives is kept in raw,
// and the ones an input reads also go to their own file here, which writes the
// record and sends the automatic replies.
package arrive

import (
	"github.com/otya0/mcbedrock-bot/internal/record/blocks"
	"github.com/otya0/mcbedrock-bot/internal/record/contents"
	"github.com/otya0/mcbedrock-bot/internal/record/names"
	"github.com/otya0/mcbedrock-bot/internal/record/raw"
	"github.com/otya0/mcbedrock-bot/internal/record/self"
	"github.com/otya0/mcbedrock-bot/internal/serverlink"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type Arrive struct {
	link     *serverlink.Link
	raw      *raw.Raw
	self     *self.Self
	contents *contents.Contents
	blocks   *blocks.Blocks
	names    *names.Names
}

// Connect sets every listener of link. Call it before link.Start.
func Connect(link *serverlink.Link, rawRecord *raw.Raw, selfRecord *self.Self, contentsRecord *contents.Contents, blocksRecord *blocks.Blocks, namesRecord *names.Names) {
	a := &Arrive{link: link, raw: rawRecord, self: selfRecord, contents: contentsRecord, blocks: blocksRecord, names: namesRecord}
	// Used by layer 2: kept in raw, then handled by the file of the same name.
	link.Connection.ChangeDimension = func(p *packet.ChangeDimension) { a.raw.Add(p); a.changeDimension(p) }
	link.Connection.NetworkStackLatency = func(p *packet.NetworkStackLatency) { a.raw.Add(p); a.networkStackLatency(p) }
	link.Dictionary.ItemRegistry = func(p *packet.ItemRegistry) { a.raw.Add(p); a.itemRegistry(p) }
	link.Dictionary.BiomeDefinitionList = func(p *packet.BiomeDefinitionList) { a.raw.Add(p); a.biomeDefinitionList(p) }
	link.Dictionary.CraftingData = func(p *packet.CraftingData) { a.raw.Add(p); a.craftingData(p) }
	link.Dictionary.CreativeContent = func(p *packet.CreativeContent) { a.raw.Add(p); a.creativeContent(p) }
	link.Terrain.NetworkChunkPublisherUpdate = func(p *packet.NetworkChunkPublisherUpdate) { a.raw.Add(p); a.networkChunkPublisherUpdate(p) }
	link.Terrain.LevelChunk = func(p *packet.LevelChunk) { a.raw.Add(p); a.levelChunk(p) }
	link.Terrain.SubChunk = func(p *packet.SubChunk) { a.raw.Add(p); a.subChunk(p) }
	link.Terrain.UpdateBlock = func(p *packet.UpdateBlock) { a.raw.Add(p); a.updateBlock(p) }
	link.Terrain.UpdateSubChunkBlocks = func(p *packet.UpdateSubChunkBlocks) { a.raw.Add(p); a.updateSubChunkBlocks(p) }
	link.Terrain.UpdateBlockSynced = func(p *packet.UpdateBlockSynced) { a.raw.Add(p); a.updateBlockSynced(p) }
	link.Entity.MovePlayer = func(p *packet.MovePlayer) { a.raw.Add(p); a.movePlayer(p) }
	link.Entity.SetActorLink = func(p *packet.SetActorLink) { a.raw.Add(p); a.setActorLink(p) }
	link.Player.DeathInfo = func(p *packet.DeathInfo) { a.raw.Add(p); a.deathInfo(p) }
	link.Player.Respawn = func(p *packet.Respawn) { a.raw.Add(p); a.respawn(p) }
	link.Operation.CorrectPlayerMovePrediction = func(p *packet.CorrectPlayerMovePrediction) { a.raw.Add(p); a.correctPlayerMovePrediction(p) }
	link.Operation.InventoryTransaction = func(p *packet.InventoryTransaction) { a.raw.Add(p); a.inventoryTransaction(p) }
	link.Operation.GUIDataPickItem = func(p *packet.GUIDataPickItem) { a.raw.Add(p); a.guiDataPickItem(p) }
	link.Operation.PlayerHotBar = func(p *packet.PlayerHotBar) { a.raw.Add(p); a.playerHotBar(p) }
	link.Inventory.ContainerOpen = func(p *packet.ContainerOpen) { a.raw.Add(p); a.containerOpen(p) }
	link.Inventory.ContainerClose = func(p *packet.ContainerClose) { a.raw.Add(p); a.containerClose(p) }
	link.Inventory.UpdateTrade = func(p *packet.UpdateTrade) { a.raw.Add(p); a.updateTrade(p) }
	link.Inventory.UpdateEquip = func(p *packet.UpdateEquip) { a.raw.Add(p); a.updateEquip(p) }
	link.Inventory.InventoryContent = func(p *packet.InventoryContent) { a.raw.Add(p); a.inventoryContent(p) }
	link.Inventory.InventorySlot = func(p *packet.InventorySlot) { a.raw.Add(p); a.inventorySlot(p) }
	link.Inventory.ItemStackResponse = func(p *packet.ItemStackResponse) { a.raw.Add(p); a.itemStackResponse(p) }
	link.Inventory.PlayerEnchantOptions = func(p *packet.PlayerEnchantOptions) { a.raw.Add(p); a.playerEnchantOptions(p) }

	// Not used by layer 2: only kept in raw, for layer 3. AddActor is also read
	// back from raw by set_actor_link, to find the runtime ID of a vehicle.
	link.Connection.NetworkSettings = func(p *packet.NetworkSettings) { a.raw.Add(p) }
	link.Connection.ServerToClientHandshake = func(p *packet.ServerToClientHandshake) { a.raw.Add(p) }
	link.Connection.ResourcePacksInfo = func(p *packet.ResourcePacksInfo) { a.raw.Add(p) }
	link.Connection.ResourcePackStack = func(p *packet.ResourcePackStack) { a.raw.Add(p) }
	link.Connection.ResourcePackDataInfo = func(p *packet.ResourcePackDataInfo) { a.raw.Add(p) }
	link.Connection.ResourcePackChunkData = func(p *packet.ResourcePackChunkData) { a.raw.Add(p) }
	link.Connection.PlayStatus = func(p *packet.PlayStatus) { a.raw.Add(p) }
	link.Connection.StartGame = func(p *packet.StartGame) { a.raw.Add(p) }
	link.Connection.ShowCredits = func(p *packet.ShowCredits) { a.raw.Add(p) }
	link.Connection.PacketViolationWarning = func(p *packet.PacketViolationWarning) { a.raw.Add(p) }
	link.Connection.Disconnect = func(p *packet.Disconnect) { a.raw.Add(p) }
	link.Connection.Transfer = func(p *packet.Transfer) { a.raw.Add(p) }
	link.Dictionary.AvailableActorIdentifiers = func(p *packet.AvailableActorIdentifiers) { a.raw.Add(p) }
	link.Dictionary.TrimData = func(p *packet.TrimData) { a.raw.Add(p) }
	link.Dictionary.VoxelShapes = func(p *packet.VoxelShapes) { a.raw.Add(p) }
	link.Dictionary.DimensionData = func(p *packet.DimensionData) { a.raw.Add(p) }
	link.Dictionary.FeatureRegistry = func(p *packet.FeatureRegistry) { a.raw.Add(p) }
	link.Dictionary.JigsawStructureData = func(p *packet.JigsawStructureData) { a.raw.Add(p) }
	link.Terrain.ChunkRadiusUpdated = func(p *packet.ChunkRadiusUpdated) { a.raw.Add(p) }
	link.Terrain.ClientCacheMissResponse = func(p *packet.ClientCacheMissResponse) { a.raw.Add(p) }
	link.BlockEntity.BlockActorData = func(p *packet.BlockActorData) { a.raw.Add(p) }
	link.BlockEntity.BlockEvent = func(p *packet.BlockEvent) { a.raw.Add(p) }
	link.BlockEntity.OpenSign = func(p *packet.OpenSign) { a.raw.Add(p) }
	link.Entity.AddActor = func(p *packet.AddActor) { a.raw.Add(p) }
	link.Entity.AddPlayer = func(p *packet.AddPlayer) { a.raw.Add(p) }
	link.Entity.AddItemActor = func(p *packet.AddItemActor) { a.raw.Add(p) }
	link.Entity.AddPainting = func(p *packet.AddPainting) { a.raw.Add(p) }
	link.Entity.RemoveActor = func(p *packet.RemoveActor) { a.raw.Add(p) }
	link.Entity.TakeItemActor = func(p *packet.TakeItemActor) { a.raw.Add(p) }
	link.Entity.MoveActorAbsolute = func(p *packet.MoveActorAbsolute) { a.raw.Add(p) }
	link.Entity.MoveActorDelta = func(p *packet.MoveActorDelta) { a.raw.Add(p) }
	link.Entity.SetActorMotion = func(p *packet.SetActorMotion) { a.raw.Add(p) }
	link.Entity.SetActorData = func(p *packet.SetActorData) { a.raw.Add(p) }
	link.Entity.UpdateAttributes = func(p *packet.UpdateAttributes) { a.raw.Add(p) }
	link.Entity.MobEffect = func(p *packet.MobEffect) { a.raw.Add(p) }
	link.Entity.MobEquipment = func(p *packet.MobEquipment) { a.raw.Add(p) }
	link.Entity.MobArmourEquipment = func(p *packet.MobArmourEquipment) { a.raw.Add(p) }
	link.Entity.Animate = func(p *packet.Animate) { a.raw.Add(p) }
	link.Entity.ActorEvent = func(p *packet.ActorEvent) { a.raw.Add(p) }
	link.Entity.PlayerSkin = func(p *packet.PlayerSkin) { a.raw.Add(p) }
	link.Player.UpdateAbilities = func(p *packet.UpdateAbilities) { a.raw.Add(p) }
	link.Player.UpdateAdventureSettings = func(p *packet.UpdateAdventureSettings) { a.raw.Add(p) }
	link.Player.SetPlayerGameType = func(p *packet.SetPlayerGameType) { a.raw.Add(p) }
	link.Player.UpdatePlayerGameType = func(p *packet.UpdatePlayerGameType) { a.raw.Add(p) }
	link.Player.SetDefaultGameType = func(p *packet.SetDefaultGameType) { a.raw.Add(p) }
	link.Player.SetSpawnPosition = func(p *packet.SetSpawnPosition) { a.raw.Add(p) }
	link.Player.SetHealth = func(p *packet.SetHealth) { a.raw.Add(p) }
	link.Player.HurtArmour = func(p *packet.HurtArmour) { a.raw.Add(p) }
	link.Player.PlayerArmourDamage = func(p *packet.PlayerArmourDamage) { a.raw.Add(p) }
	link.Player.UpdateClientInputLocks = func(p *packet.UpdateClientInputLocks) { a.raw.Add(p) }
	link.Player.MovementEffect = func(p *packet.MovementEffect) { a.raw.Add(p) }
	link.Operation.ServerPlayerPostMovePosition = func(p *packet.ServerPlayerPostMovePosition) { a.raw.Add(p) }
	link.Operation.CompletedUsingItem = func(p *packet.CompletedUsingItem) { a.raw.Add(p) }
	link.Operation.ClientStartItemCooldown = func(p *packet.ClientStartItemCooldown) { a.raw.Add(p) }
	link.Operation.Emote = func(p *packet.Emote) { a.raw.Add(p) }
	link.Operation.EmoteList = func(p *packet.EmoteList) { a.raw.Add(p) }
	link.Inventory.ContainerSetData = func(p *packet.ContainerSetData) { a.raw.Add(p) }
	link.Inventory.SetPlayerFurnaceOptions = func(p *packet.SetPlayerFurnaceOptions) { a.raw.Add(p) }
	link.Inventory.SetPlayerInventoryOptions = func(p *packet.SetPlayerInventoryOptions) { a.raw.Add(p) }
	link.Inventory.ContainerRegistryCleanup = func(p *packet.ContainerRegistryCleanup) { a.raw.Add(p) }
	link.Inventory.UnlockedRecipes = func(p *packet.UnlockedRecipes) { a.raw.Add(p) }
	link.Text.Text = func(p *packet.Text) { a.raw.Add(p) }
	link.Display.SetTitle = func(p *packet.SetTitle) { a.raw.Add(p) }
	link.Display.BossEvent = func(p *packet.BossEvent) { a.raw.Add(p) }
	link.Display.SetDisplayObjective = func(p *packet.SetDisplayObjective) { a.raw.Add(p) }
	link.Display.SetScore = func(p *packet.SetScore) { a.raw.Add(p) }
	link.Display.SetScoreboardIdentity = func(p *packet.SetScoreboardIdentity) { a.raw.Add(p) }
	link.Display.RemoveObjective = func(p *packet.RemoveObjective) { a.raw.Add(p) }
	link.Display.PlayerList = func(p *packet.PlayerList) { a.raw.Add(p) }
	link.Display.PlayerLocation = func(p *packet.PlayerLocation) { a.raw.Add(p) }
	link.Display.LocatorBar = func(p *packet.LocatorBar) { a.raw.Add(p) }
	link.Environment.LevelEvent = func(p *packet.LevelEvent) { a.raw.Add(p) }
	link.Environment.LevelSoundEvent = func(p *packet.LevelSoundEvent) { a.raw.Add(p) }
	link.Environment.SetDifficulty = func(p *packet.SetDifficulty) { a.raw.Add(p) }
	link.Environment.SyncWorldClocks = func(p *packet.SyncWorldClocks) { a.raw.Add(p) }
	link.Environment.SetTime = func(p *packet.SetTime) { a.raw.Add(p) }
	link.Environment.GameRulesChanged = func(p *packet.GameRulesChanged) { a.raw.Add(p) }
	link.Environment.CurrentStructureFeature = func(p *packet.CurrentStructureFeature) { a.raw.Add(p) }
	link.Map.ClientBoundMapItemData = func(p *packet.ClientBoundMapItemData) { a.raw.Add(p) }
	link.Map.PositionTrackingDBServerBroadcast = func(p *packet.PositionTrackingDBServerBroadcast) { a.raw.Add(p) }
	link.OtherPacket = a.raw.Add
}
